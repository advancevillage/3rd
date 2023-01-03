package rs

import (
	"github.com/advancevillage/3rd/mathx/gf"
)

type Rs interface {
}

// Reed-Solomon(RS) 里德-所罗门编码； 该编码定义在伽罗瓦域中，将数据流重新排列，以块(symbol)为单位。
//
//
//    ^      <--------------------- n symbol  RS 编码--------------------->
//    .      | | | | ........................................| | |....| | |
//    .      |-| | | ........................................| | |....| | |
//    m      |-| | | ........................................| | |....| | |
//   bits    |-| | | ........................................| | |....| | |
//    .      |-| | | ........................................| | |....| | |
//    .      |-| | | ........................................| | |....| | |
//    .      |-| | | ........................................| | |....| | |
//    v      |_| | | ........................................| | |....| | |
//           <---------------- k symbol 原数据--------------->
//                                                           <-- n-k=2t -->
//                                                               校验数据
//
// 编码原理:
//
// 1. 将数据看着多项式，用数组表示 array[0] 表示 x^0 系数是 array[0];  array[1] 表示 x^1 的系数是 array[1] .....
//
//    即：
//       m(x) = m[k-1] * x^(k-1) + m[k-2] * x^(k-2) + ..... m[1] * x^1 + m[0]
//
// 2. 生成纠错多项式g(x)
//
//	  即：
//             2t - 1
//	           ------
//		        | |
//		 g(x) = | |    (x - a^j)   = (x - a^0)(x - a^1)(x - a^2)(x - a^(2t-1))
//              | |
//	           ------
//             j = 0
//
//
// 3. IEEE 802.3 119.2.4.6中定义了编码的算法：
//
//    注意： n = k + 2t
//
//             m(x) * x^2t  % g(x)  =  m[k-1] * x^(n-1) + m[k-2] * x^(n-2) + .... + m0 * x^2t + 0 * x^(2t-1) + ..... + 0 * x^0
//																															   % g(x)
//                                  =  p[2t-1] * x^(2t-1) + p[2t-2] * x^(2t-2) + ... + p[0]
//
//
// 所以RS编码结果是
//
//            m[k-1] * x^(n-1) + m[k-2] * x^(n-2) + .... + m[0] * x^(2t) + p[2t-1] * x^(2t-1) + p[2t-2] * x^(2t-2) + ... + p[0]
//
//   即：
//        m[k-1] m[k-2] m[k-3] ... m[0] p[2t-1] p[2t-2] ... p[0]
//

type rs struct {
	g   gf.Gf
	gop gf.GfOp
}

func newrs(m uint32) (*rs, error) {
	var r = &rs{}

	g, err := gf.NewGf(m)
	if err != nil {
		return nil, err
	}

	gop, err := gf.NewGfOp(g)
	if err != nil {
		return nil, err
	}

	r.g = g
	r.gop = gop

	return r, nil
}

// number of error correction codewords
func (r *rs) ggx(nEcCw uint32) []uint32 {

	p := []uint32{1}

	for i := uint32(0); i < nEcCw; i++ {
		var (
			pp     = []uint32{r.g.Atop(i), 1}
			matrix = make([][]uint32, len(pp))
		)

		for j := 0; j < len(matrix); j++ {
			matrix[j] = make([]uint32, len(p))
		}

		for j := 0; j < len(pp); j++ {
			for k := 0; k < len(p); k++ {
				matrix[j][k] = r.gop.Mul(pp[j], p[k], gf.OpPloy)
			}
		}

		for j := 0; j < len(p); j++ {
			p[j] = 0
		}
		p = append(p, 0)

		for j := 0; j < len(pp); j++ {
			for k := 0; k < len(matrix[j]); k++ {
				p[j+k] = r.gop.Add(p[j+k], matrix[j][k], gf.OpPloy)
			}
		}

	}

	return p
}

func (r *rs) encode(cw []uint32, nEcCw uint32) []uint32 {
	// 1. 生成纠错多项式
	gx := r.ggx(nEcCw)

	ngx := len(gx) - 1
	ncw := len(cw)

	// 2. 生成数据多项式
	mx := make([]uint32, ncw+ngx)

	for i := 0; i < ncw; i++ {
		mx[ngx+i] = cw[i]
	}

	// 3. 多项式相除
	dx := make([]uint32, ncw+ngx)

	for i := ncw - 1; i >= 0; i-- {

		// 高阶系数是0转下一位处理
		if mx[i+ngx] == 0 {
			continue
		}
		// 阶差、系数
		coef := r.gop.Div(mx[i+ngx], gx[ngx], gf.OpPloy)

		// 消元
		for j := 0; j <= ngx; j++ {
			dx[i+j] = r.gop.Mul(gx[j], coef, gf.OpPloy)
		}

		for j := 0; j < ngx+ncw; j++ {
			mx[j] = r.gop.Add(mx[j], dx[j], gf.OpPloy)
		}

		// 重置
		for j := 0; j < ngx+ncw; j++ {
			dx[j] = 0
		}
	}

	for i := 0; i < ncw; i++ {
		mx[ngx+i] = cw[i]
	}

	return mx
}
