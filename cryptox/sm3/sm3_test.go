package sm3

import (
	"bytes"
	"encoding/hex"
	"testing"
)

var sm3Test = map[string]struct {
	m []byte
	e string
}{
	"case1": {
		m: []byte{0x61, 0x62, 0x63},
		e: "66C7F0F462EEEDD9D1F2D46BDC10E4E24167C4875CF2F7A2297DA02B8F4BA8E0",
	},
	"case2": {
		m: []byte{
			0x61, 0x62, 0x63, 0x64, 0x61, 0x62, 0x63, 0x64, 0x61, 0x62, 0x63, 0x64, 0x61, 0x62, 0x63, 0x64,
			0x61, 0x62, 0x63, 0x64, 0x61, 0x62, 0x63, 0x64, 0x61, 0x62, 0x63, 0x64, 0x61, 0x62, 0x63, 0x64,
			0x61, 0x62, 0x63, 0x64, 0x61, 0x62, 0x63, 0x64, 0x61, 0x62, 0x63, 0x64, 0x61, 0x62, 0x63, 0x64,
			0x61, 0x62, 0x63, 0x64, 0x61, 0x62, 0x63, 0x64, 0x61, 0x62, 0x63, 0x64, 0x61, 0x62, 0x63, 0x64,
		},
		e: "DEBE9FF92275B8A138604889C18E5A4D6FDB70E5387E5765293DCBA39C0C5732",
	},
	"case3": {
		m: []byte("Discard medicine more than two years old."),
		e: "8A89BD24087AE6F9A3AAE485BFA9ECD276F909A04B248EAB1B4F9BE2B24F0111",
	},
	"case4": {
		m: []byte("He who has a shady past knows that nice guys finish last."),
		e: "2BB6C53AD20EAF2552425F44E72D96D1B61E63310A1A30F4E5406A103619177D",
	},
	"case5": {
		m: []byte("I wouldn't marry him with a ten foot pole."),
		e: "5ECEC640017AFD77D00147EF42FDB8E7901F089A62C1888637917E89BB3A6532",
	},
	"case6": {
		m: []byte("The fugacity of a constituent in a mixture of gases at a given temperature is proportional to its mole fraction."),
		e: "87E709CF62ACADCA93B8012483041BA7113446285E6FC20DAE868FC0557A2CC5",
	},
}

func Test_sm3(t *testing.T) {
	for n, p := range sm3Test {
		f := func(t *testing.T) {
			// write append
			h1 := New()
			for i := range p.m {
				n, err := h1.Write([]byte{p.m[i]})
				if n != 1 || err != nil {
					t.Fatal("write bytes not equal 1")
					return
				}
			}
			var e1 = h1.Sum(nil)

			// write ont time
			var e2 []byte
			sign := Sum(p.m)
			e2 = append(e2, sign[:]...)

			// check
			if !bytes.Equal(e1, e2) {
				t.Fatalf("e1(%s) not equal e2(%s)", hex.EncodeToString(e1), hex.EncodeToString(e2))
				return
			}

			e3, err := hex.DecodeString(p.e)
			if err != nil {
				t.Fatalf("p.e not hex string")
				return
			}
			if !bytes.Equal(e2, e3) {
				t.Fatalf("e2(%s) not equal e3(%s)", hex.EncodeToString(e2), hex.EncodeToString(e3))
				return
			}
		}
		t.Run(n, f)
	}
}
