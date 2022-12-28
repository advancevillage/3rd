package bch

import (
	"fmt"

	"github.com/advancevillage/3rd/mathx/gf"
)

type Bch interface {
	Encode(m uint32) uint32
}

func NewPrimitiveBch(m, t uint32) (Bch, error) {
	var b, err = newbch(m, t)
	if err != nil {
		return nil, err
	}
	b.gmp()
	b.ggx()
	return b, nil
}

func NewBch(n, gx uint32) (Bch, error) {
	var b = &bch{}

	b.n = n

	nn := 0 // 检查位长度
	// g(x) = x^12 + x^11 + x^10 + x^9 + x^8 + x^5 + x^2 + 1
	// g(x) = 0b1 1111 0010 0101
	// len(gx) = 12
	for i := 0; i < 32; i++ {
		if (gx>>i)&0x1 > 0 {
			nn = i
		} else {
			continue
		}
	}
	b.k = n - uint32(nn)        // 数据位长度
	b.gx = make([]uint32, nn+1) // 生成多项式

	for i := 0; i <= nn; i++ {
		b.gx[i] = (gx >> i) & 0x1
	}

	return b, nil
}

type bch struct {
	g   gf.Gf
	gop gf.GfOp
	m   uint32     // GF(2^m)
	n   uint32     // 2^m
	t   uint32     // 纠错个数
	k   uint32     // 数据有效位
	mp  [][]uint32 // 极小多项式表
	gx  []uint32   // 生成多项式
}

func newbch(m, t uint32) (*bch, error) {
	var c = new(bch)
	g, err := gf.NewGf(m)
	if err != nil {
		return nil, err
	}

	gop, err := gf.NewGfOp(g)
	if err != nil {
		return nil, err
	}

	c.g = g
	c.gop = gop
	c.t = t
	c.n = 1 << m
	c.m = m

	if c.t<<1 >= c.n {
		return c, fmt.Errorf("param t is too big; should be small %d", c.n>>1)
	}

	c.mp = make([][]uint32, c.n)
	for i := uint32(0); i < c.n; i++ {
		c.mp[i] = make([]uint32, 0, c.m)
	}

	return c, nil
}

// Minimal Polynomial；最小多项式, 根据伽罗瓦定理：
//
// beta是GF(q^m)中的元素，以beta为根的阶数最小多项式，称为最小多项式F(X)，即 F(beta) = 0
//
//         e-1
//         __
//  F(X) = ||  (X - Beta^(q^i)
//         --
//         i=0
//
// 满足 beta^(q^e) = beta, e为最小正整数; 即 beta^(q^e - 1) = 1
//
// 对于BCH编码, q = 2, e = m
//
// GMP (Generate Minimal Polynomial)，生成极小多项式表 (x - a^(2^0))(x - a^(2^1))(x - a^(2^2))(x - a^(2^3)) m = 4;
func (c *bch) gmp() {
	var table = make([]bool, c.n+1)

	var i, j, k uint32

	for i < c.n {
		c.mp[i] = append(c.mp[i], i) // 以a^i次幂为根

		// 根据伽罗瓦定理3：
		//    如果f(X)是GF(p)上的多项式，beta是GF(p^m)中的元素，如果beta是f(X)的一个根，那么对于任意正整数t，
		// beta^(p^t)也是f(X)的一个根。beta^(p^t)称为beta的共轭
		//
		// 例如 m = 4
		//  i |
		// ----------------------------
		//  0 |  0
		//  1 |  1  2  4  8
		//  2 |  1  2  4  8
		//  3 |  3  6 12  9
		//  4 |  4  8  1  2
		//  5 |  5 10
		//  6 |  6 12  9  3
		//  7 |  7 14 13 11
		//  8 |  8  1  2  4
		//  9 |  9  3  6 12
		// 10 | 10  5
		// 11 | 11  7 14 13
		// 12 | 12  9  3  6
		// 13 | 13 11  7 14
		// 14 | 14 13 11  7
		// 15 | 15
		//
		for j = 1; j < c.n; j++ { // 寻找 a^i 的共轭根
			if (i*(1<<j))%(c.n-1) == i%(c.n-1) {
				break
			} else {
				c.mp[i] = append(c.mp[i], (i*1<<j)%(c.n-1))
			}
		}

		for k = 0; k < uint32(len(c.mp[i])); k++ {
			table[c.mp[i][k]] = true
			c.mp[c.mp[i][k]] = c.mp[i]
		}

		for k = 0; k < uint32(len(table)); k++ {
			if table[k] {
				continue
			} else {
				i = k
				break
			}
		}
	}
}

// BCH码生成多项式
//      g[x] = LCM(f1[x]f2[x]f3[x]......f(2t)[x])
//
// f(2t)[x] 是极小多项式
//
// 涉及多项式相乘，多项式相乘用矩阵求解
//
//                      |A|             |AB   A|
//  (x + A)(x + B) =    | | x |B  1| =  |      |  = AB + (A+B)x + x^2
//                      |1|             |B    1|
//
// 注意：AB 为伽罗瓦乘法运算， A + B 为伽罗瓦加法运算
//
//                                  |AB |                | ABC     AB   0|
//  (AB + (A+B)x + x^2)(C + x) =    |A+B| x |C  1  0| =  |(A+B)C   A+B  0|  = ABC + [(A+B)C + AB]x + (A+B+C)x^2 + x^3
//                                  | 1 |                |  C       1   0|
//
func (c *bch) ggx() {
	//1. 计算生成多项式的极小多项式列表
	var (
		alpha  []uint32
		malpha = make(map[uint32]bool)
	)

	for i := uint32(1); i <= (c.t << 1); i++ {

		for j := 0; j < len(c.mp[i]); j++ {

			if malpha[c.mp[i][j]] {
				continue
			} else {
				alpha = append(alpha, c.mp[i][j])
				malpha[c.mp[i][j]] = true
			}

		}
	}

	//2. 生成多项式
	p := []uint32{1}

	for i := 0; i < len(alpha); i++ {
		var (
			pp     = []uint32{c.g.Atop(alpha[i]), 1}
			matrix = make([][]uint32, len(pp))
		)

		for j := 0; j < len(matrix); j++ {
			matrix[j] = make([]uint32, len(p))
		}

		for j := 0; j < len(pp); j++ {
			for k := 0; k < len(p); k++ {
				matrix[j][k] = c.gop.Mul(pp[j], p[k], gf.OpPloy)
			}
		}

		for j := 0; j < len(p); j++ {
			p[j] = 0
		}
		p = append(p, 0)

		for j := 0; j < len(pp); j++ {
			for k := 0; k < len(matrix[j]); k++ {
				p[j+k] = c.gop.Add(p[j+k], matrix[j][k], gf.OpPloy)
			}
		}
	}

	c.gx = p
	c.k = c.n - uint32(len(c.gx))
}

// BCH是循环码的一种，BCH(n,k)中n编码长度，k信息长度，信息码多项式m(x)
//
//  [x^(n - k) * m(x)] / gx = Q(x) ... r(x)
//
// 即  x^(n - k) * m(x) = Q(x)*g(x) + r(x)
//
// C(x) = x^(n - k) * m(x) + r(x)
//
// encode_bch: http://www.eccpage.com/bch3.c
func (c *bch) encode(m uint32) uint32 {
	var (
		n  = int(c.k)
		nn = len(c.gx) - 1
		mx = make([]uint32, nn+n)
	)

	for i := 0; i < n; i++ {
		mx[nn+i] = (m >> i) & 0x1
	}

	for i := n - 1; i >= 0; i-- {
		feedback := mx[nn+i] ^ mx[nn-1]

		if feedback > 0 {
			for j := nn - 1; j > 0; j-- {
				if c.gx[j] > 0 {
					mx[j] = mx[j-1] ^ feedback
				} else {
					mx[j] = mx[j-1]
				}
			}
			mx[0] = c.gx[0] & feedback
		} else {
			for j := nn - 1; j > 0; j-- {
				mx[j] = mx[j-1]
			}
			mx[0] = 0
		}
	}

	m = 0

	for i := 0; i < nn+n; i++ {
		m |= mx[i] << i
	}

	return m
}

func (c *bch) Encode(m uint32) uint32 {
	return c.encode(m)
}
