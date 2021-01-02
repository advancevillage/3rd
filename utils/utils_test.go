package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var jwtTestData = map[string]struct {
	sct    string
	except bool
	exp    int
	sm     SignMethod
}{
	"case1": {
		sct:    RandsString(5),
		exp:    5,
		except: true,
		sm:     HS256,
	},
	"case2": {
		sct:    RandsString(5),
		exp:    5,
		except: true,
		sm:     HS384,
	},
	"case3": {
		sct:    RandsString(5),
		exp:    5,
		except: true,
		sm:     HS512,
	},
	"case4": {
		sct:    RandsString(5),
		exp:    5,
		except: true,
		sm:     PS256,
	},
	"case5": {
		sct:    RandsString(5),
		exp:    5,
		except: true,
		sm:     PS384,
	},
	"case6": {
		sct:    RandsString(5),
		exp:    5,
		except: true,
		sm:     PS512,
	},
	"case7": {
		sct:    RandsString(5),
		exp:    5,
		except: true,
		sm:     ES256,
	},
	"case8": {
		sct:    RandsString(5),
		exp:    5,
		except: true,
		sm:     ES384,
	},
	"case9": {
		sct:    RandsString(5),
		exp:    5,
		except: true,
		sm:     ES512,
	},
}

func Test_jwtClient(t *testing.T) {
	for n, p := range jwtTestData {
		f := func(t *testing.T) {
			var (
				c     ITokenClient
				e     error
				b     bool
				token string
			)
			c, e = NewJwtClient(p.sct, p.sm)
			if e != nil {
				t.Fatal(e)
				return
			}
			token, e = c.CreateToken(p.exp)
			if e != nil {
				t.Fatal(e)
				return
			}
			fmt.Println(n, p.sct, token)
			b, e = c.ParseToken(token)
			if e != nil {
				t.Fatal(e)
				return
			}
			assert.Equal(t, p.except, b)
		}
		t.Run(n, f)
	}
}
