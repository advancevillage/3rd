package sm3

import "encoding/binary"

func blockGeneric(dig *digest, p []byte) {
	//load state
	a, b, c, d, e, f, g, h := dig.s[0], dig.s[1], dig.s[2], dig.s[3], dig.s[4], dig.s[5], dig.s[6], dig.s[7]

	//GB/T 32905-2016
	for i := 0; i <= len(p)-BlockSize; i += BlockSize {
		//GB/T 32905-2016 5.3.2. Message Extension
		q := p[i:]
		q = q[:BlockSize:BlockSize]

		w := [132]uint32{}

		for j := 0; j < 132; j++ {
			switch {
			case j < 16:
				w[j] = binary.BigEndian.Uint32(q[4*j:])

			case j < 68:
				w[j] = p1(w[j-16]^w[j-9]^rol32(w[j-3], 15)) ^ rol32(w[j-13], 7) ^ w[j-6]

			default:
				w[j] = w[j-68] ^ w[j-64]
			}
		}

		for j := 0; j < 64; j++ {
			ss1 := rol32(rol32(a, 12)+e+rol32(kk(j), j%32), 7)
			ss2 := ss1 ^ rol32(a, 12)
			tt1 := ff(j, a, b, c) + d + ss2 + w[68+j]
			tt2 := gg(j, e, f, g) + h + ss1 + w[j]
			d = c
			c = rol32(b, 9)
			b = a
			a = tt1
			h = g
			g = rol32(f, 19)
			f = e
			e = p0(tt2)
		}

		dig.s[0] ^= a
		dig.s[1] ^= b
		dig.s[2] ^= c
		dig.s[3] ^= d
		dig.s[4] ^= e
		dig.s[5] ^= f
		dig.s[6] ^= g
		dig.s[7] ^= h
	}
}
