package utils

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"regexp"
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
		sm:     RS256,
		except: true,
		exp:    5,
		sct: `
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAwT5YWfpE3HvJ0Q9a1U5dRdV36hIMdgqD5k/AaXCwIhWrquS/
6Qd/J14i9EWq9dTUiuEMXcSxFxmNnLW+fPDzbDUXkJ78h+ZGDfSEoktfozPGXUvW
ZjGX21e+GYt3WShoviZTmd4a4FG9gHMa/VdbpaAfl/qFV2IAoRyHDf8B5kED+FtQ
3W27zuTUNCkgZXfWif9qCdUrotpbofMx001VW/QneQaLuC7fP/QdRaFBwnAR2L2u
vbiZ3IrUtI6cVZkFmGDExsvsOLvY+90SyZwQnchCkXUP9IEErPlatGwqFeFKPx68
Qs4tJZ3sJ6UMXf5/opuwjn21TVTRnZ5EPonKPQIDAQABAoIBAHo54NAoh92dPm1I
9P7t7fj5qDsY52DSDdNipiUK7ZzhcA9LWEcgQsC3vgwa9KNA1p5w8c1tV0VxGC0C
l1WXYaAThLAonzml4LF681ljqz4ixVjFWvqQa6iEuuyVVgvCj12WCFLONNmlWeMg
6vVKh+Eegl0yS5yVlChTTuj/XkgvxdvSNkiiyZMiT/06l9s+Vlq7YJT2d76DMv78
djGAzXzmXbkntHn2JGGMH58B0SEulY4VRFaNCTznRjerz/wUvKmzu3jmGEauVZhz
r2rp58qcQxilEVWqFVqJuvxKlWx5TSnhvZcF4mw4DlE9HfsllCPmVOOYugU2ITJM
7JOqXgECgYEA4VjSGUHhvn0wDTKdZFad4td8uO+q1KQQm50z7pb0WyQ/kESYfDi7
elYtnal1qQioDzOl0HWCneXYdPxSX7XHRn6SWoxp/NLTNm31PNu+zef2EVcSoZwz
625SXrmGzyHH+DbOKXrFS/zZsUCgbMpZzAO4P30CuRCsZksDW8/cdl0CgYEA24eZ
HPW7W1P0cW/OJ63HVyu4AXW4uaEeq6yb9FJD/GGHareYcPd3wJBx+zsHNTPRNS8E
1PIDPrmA7GL4xjjgV25J/4fd9ohUN0zansaA+idOvFObA5XQ3zaBgAGt/BL2DwrR
CJx8brh/4prcgpLOYWYnizFDd5GDQCnU5LkjpWECgYEAz+WPv3mGeKUqJaLijeVT
OCoxiLSj2BWctNQtSxq9STCB6+k8/K2iWNUUtHXYdR/lXFD70vv2ixG3xwXaQS6F
MEYpY5xTU8p0zaxYKdNZjsFHxOud5rcjNzrKh1WGR6YUxKxbOu1nBBm8BMlot7Vf
btanrvr3/iChzKW77YIxFIkCgYBRTrKd8EF5POoPTZqsRYfMOGwJVmGZlxw191M3
tXRquHCgTOmQBYf78UPWCfHMeamlqgl/GTesdqZSZwG+4PfuSfHsS5UhJzMR3Ewo
fFruy7o0tD54oHdhBN4H3BdlglxSC+6J4vOPSpRLCJJdZiQ8HMrOmetkEKftDtFD
+XJDYQKBgBj/V8OPdYiCV8vIBqqP374pzVFOrZVNnvecbLN1X5dVvWajPmmyY3E5
LKtS41UPWaL8d2v7G4kXaIUEeUyAV4WYlTF4lHDwj+YzmwFF0jGNKEKaWRAl4WC+
20gRPVPhBuRHzf2fBj+SmA2eSvt5018dxlP6kkX2ICy3HMlXHlj5
-----END RSA PRIVATE KEY-----
|
-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAwT5YWfpE3HvJ0Q9a1U5d
RdV36hIMdgqD5k/AaXCwIhWrquS/6Qd/J14i9EWq9dTUiuEMXcSxFxmNnLW+fPDz
bDUXkJ78h+ZGDfSEoktfozPGXUvWZjGX21e+GYt3WShoviZTmd4a4FG9gHMa/Vdb
paAfl/qFV2IAoRyHDf8B5kED+FtQ3W27zuTUNCkgZXfWif9qCdUrotpbofMx001V
W/QneQaLuC7fP/QdRaFBwnAR2L2uvbiZ3IrUtI6cVZkFmGDExsvsOLvY+90SyZwQ
nchCkXUP9IEErPlatGwqFeFKPx68Qs4tJZ3sJ6UMXf5/opuwjn21TVTRnZ5EPonK
PQIDAQAB
-----END PUBLIC KEY-----
`,
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

var uuidTestData = map[string]struct {
	except string
	count  int
}{
	"case1": {
		except: `\b[0-9a-f]{8}\b-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-\b[0-9a-f]{12}\b`,
		count:  10000,
	},
}

func Test_uuid(t *testing.T) {
	for n, p := range uuidTestData {
		f := func(t *testing.T) {
			var valid = regexp.MustCompile(p.except)
			for i := 0; i < p.count; i++ {
				var uuid = UUID()
				if !valid.MatchString(uuid) {
					t.Fatal(uuid)
					return
				}
			}
		}
		t.Run(n, f)
	}
}

var uuid8TestData = map[string]struct {
	except int
	con    int
}{
	"case1": {
		except: 1000,
		con:    1000,
	},
	"case2": {
		except: 10000,
		con:    10000,
	},
	"case3": {
		except: 20000,
		con:    20000,
	},
	"case4": {
		except: 30000,
		con:    30000,
	},
	"case5": {
		except: 40000,
		con:    40000,
	},
}

func Test_uuid8(t *testing.T) {
	for n, p := range uuid8TestData {
		f := func(t *testing.T) {
			var m = make(map[string]struct{})
			for i := 0; i < p.con; i++ {
				var k = string(UUID8Byte())
				m[k] = struct{}{}
			}
			assert.Equal(t, p.except, len(m))
		}
		t.Run(n, f)
	}
}

var secp256r1TestData = map[string]struct {
	key    []byte
	except []byte
}{
	"case1": {},
}

func Test_secp256r1(t *testing.T) {
	for n, p := range secp256r1TestData {
		f := func(t *testing.T) {
			var curve = elliptic.P256()
			var iPri, err = ecdsa.GenerateKey(curve, rand.Reader)
			if err != nil {
				t.Fatal(err)
				return
			}
			var iPub = append(iPri.PublicKey.X.Bytes(), iPri.PublicKey.Y.Bytes()...)
			fmt.Printf("iPri %x\n", iPri.D.Bytes())
			fmt.Printf("iPub %x\n", iPub)
			fmt.Println(p)
		}
		t.Run(n, f)
	}
}
