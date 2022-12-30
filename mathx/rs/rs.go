package rs

import "github.com/advancevillage/3rd/mathx/gf"

type Rs interface {
}

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
