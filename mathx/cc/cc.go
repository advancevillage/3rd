package cc

import "fmt"

// 循环码
// 设C为(n,k)线性分组码的码组集合。对C中任意码组c定义
//
//     c = (a(n-1) a(n-2) ...... a0)  c ∈C
//
// 循环移位1位
//     c¹= (a(n-2) a(n-3) ...... a0 a(n-1))  c¹∈C
// 循环移位2位
//     c²= (a(n-3) a(n-4) ...... a0 a(n-1) a(n-2))  c²∈C
// 循环移位i位
//     cⁱ = (a(n-i-1)) a(n-i-2) ...... a0 a(n-1) a(n-2) a(n-i))  cⁱ ∈C
//
//-----------------------------------------------------------------------
// 码多项式 表示循环码
//
//     c = (a(n-1) a(n-2) ...... a0)  c ∈C
// 用多项式表示
//
//	   c(x) = a(n-1)x⁽ⁿ⁻¹⁾ + a(n-2)x⁽ⁿ⁻²⁾ + .... + a₀x⁰
// a₍ₙ₋
//
// eg: 某码组1100101 用多项式表示
//     c₇(x) = 1·x⁶ + 1·x⁵ + 0·x⁴+ 0·x³ + 1·x²+ 0·x¹+ 1
//           = x⁶ + x⁵ + x²+ 1
//
// 码多项式的模运算
// 若正整数M除以正整数N，商为Q, 余数为R，可以表示为
//
//			M / N = Q + R/N ( 0 <= R < N)
// 记 M ≅ R
//
// 两个多项式a(x) 和 p(x) , 一定存在唯一的多项式Q(x)和r(x)，使得：
//          a(x) = Q(x)·p(x) + r(x)
//
// 记 a(x) ≅ r(x)  [mod p(x)], 0 <= deg(r(x)) < deg(p(x))
//
// 定理:
//	   对于(n,k)循环码，若c(x) = (a(n-1) a(n-2) .... a₁a₀)，则有：
//					cⁱ(x) = [xⁱc(x)]%(xⁿ+ 1)
//-----------------------------------------------------------------------
// 生成多项式
//	g(x) 是 (n-k) 次多项式，所以xᵏg(x)是n次多项式，则：
//
//		xᵏg(x) / xⁿ+ 1 = 1 ..... b(x)
//
// 即:
//		xᵏg(x) = 1·(xⁿ+ 1) + b(x)
//表示g(x)左移k位，并且仍是许用码组，即是g(x)倍式，则：
//
//		b(x) = u(x)g(x)
//
//即证明g(x)是xⁿ+ 1的一个因子。这个性质指出了g(x)的求解方法，即对多项式xⁿ+ 1进行因式分解
//
//示例:
//	 (7,k)循环码的生成多项式，x⁷+1多项式分解
//			(x+1)(x³+ x²+ 1)(x³+ x + 1)
//
type ICyclicCode interface {
	G() [][]byte
	H() [][]byte
}

var _ ICyclicCode = cc{}

type cc struct {
	n       int      // 码组
	k       int      // 信息位
	gx      []byte   // 生成多项式
	MatrixG [][]byte // 生成矩阵
	MatrixH [][]byte // 监督矩阵
}

// 基于(n,k) 构建循环码

// 基于(n,gx)构建循环码; deg(gx) < n
// 参数:
//		gx  生成多项式，[]byte. g(x) = gx[0]·x⁰ + gx[1]·x¹+ gx[2]·x²..... + gx[i]·xⁱ
//
// deg(g): 非零系数的最高次幂
//
// gx = gx[0] + gx[1]x¹+gx[2]x² .... +gx[i]xⁱ
//
func NewWithGx(n int, gx []byte) (ICyclicCode, error) {
	switch {
	case len(gx) == 0:
		return nil, fmt.Errorf("gx len is %d", len(gx))

	case len(gx) >= n:
		return nil, fmt.Errorf("gx len is %d, but n is %d", len(gx), n)
	}

	var c = cc{
		n:  n,
		k:  n - len(gx) + 1,
		gx: gx,
	}

	// k位信息多项式 u(x) = uₖ₋₁·xᵏ⁻¹+ uₖ₋₂·xᵏ⁻²+ uₖ₋₃·xᵏ⁻³+ uₖ₋₄·xᵏ⁻⁴+ ... + u₀
	// n位编码多项式 c(x) = xⁿ⁻ᵏ·u(x) + r(x)
	//					  = uₖ₋₁·xⁿ⁻¹+ uₖ₋₂·xⁿ⁻²+ uₖ₋₃·xⁿ⁻³+ uₖ₋₄·xⁿ⁻⁴+ ... + u₀ ·xⁿ⁻ᵏ + r(x)
	//
	// [xⁿ⁻ᵏ ·u(x) ]%g(x) = [uₖ₋₁·xⁿ⁻¹]%g(x) + [uₖ₋₂·xⁿ⁻²]%g(x) + [uₖ₋₃·xⁿ⁻³]%g(x) + [uₖ₋₄·xⁿ⁻⁴]%g(x) + ... + [u₀ ·xⁿ⁻ᵏ]%g(x)
	//
	//                                                      [ xⁿ⁻¹   % g(x) ]
	//                                                      [ xⁿ⁻²   % g(x) ]
	//						= [uₖ₋₁uₖ₋₂uₖ₋₃uₖ₋₄ ... u₀] [ ....            ] = r(x) = Q
	//                                                      [ xⁿ⁻ᵏ⁻¹ % g(x) ]
	//                                                      [ xⁿ⁻ᵏ    % g(x) ]
	//
	//                                       [ xⁿ⁻¹  ]
	//                                       [ xⁿ⁻²  ]
	// U ·Iₖ = [uₖ₋₁uₖ₋₂uₖ₋₃uₖ₋₄ ... u₀][ ....    ] =  xⁿ⁻ᵏ ·u(x)
	//                                       [ xⁿ⁻ᵏ⁻¹]
	//                                       [ xⁿ⁻ᵏ   ]
	// G = [Iₖₓₖ|Qₖₓ₍ₙ₋ₖ₎]
	//
	c.MatrixG = make([][]byte, c.k)
	for i := 0; i < c.k; i++ {
		c.MatrixG[i] = c.gG(c.n - 1 - i)
	}

	// H = [Qₖₓ₍ₙ₋ₖ₎ᵀ|I₍ₙ₋ₖ₎ₓ₍ₙ₋ₖ₎]
	c.MatrixH = make([][]byte, c.n-c.k)
	for i := 0; i < c.n-c.k; i++ {
		c.MatrixH[i] = c.gH(i)
	}
	return c, nil
}

func (c cc) gG(pos int) []byte {
	var (
		p   = make([]byte, c.n)
		deg = len(c.gx) - 1
	)
	p[pos] = 1

	for i := c.n - 1; i >= deg; i-- {

		if p[i] <= 0 {
			continue
		}

		for j := deg; j >= 0; j-- {
			p[i+j-deg] ^= c.gx[j]
		}
	}

	p[pos] = 1
	return p
}

func (c cc) gH(pos int) []byte {
	var (
		p = make([]byte, c.n)
		n = c.n - c.k - 1
	)
	p[n-pos] = 1

	for i := 0; i < c.k; i++ {
		p[c.n-1-i] = c.MatrixG[i][n-pos]
	}

	return p
}

func (c cc) G() [][]byte {
	return c.MatrixG
}

func (c cc) H() [][]byte {
	return c.MatrixH
}
