package x_test

import (
	"testing"

	"github.com/advancevillage/3rd/x"
	"github.com/stretchr/testify/assert"
)

func Test_Cfg(t *testing.T) {
	var data = map[string]struct {
		filepath string
		cfg      *x.Cfg
	}{
		"case1": {
			filepath: "tcfg.yaml",
			cfg: &x.Cfg{
				App: &x.AppCfg{
					Name: "xmagic",
					Host: "127.0.0.1",
					Port: 1995,
				},
				Log: &x.LogCfg{
					Level: "debug",
				},
				Creds: &x.CredCfg{
					MTls: &x.MTlsCfg{
						Cert: "cert.pem",
						Key:  "private.pem",
					},
				},
			},
		},
	}

	for n, v := range data {
		f := func(t *testing.T) {
			cfg, err := x.NewCfg(v.filepath)
			assert.Nil(t, err)
			assert.Equal(t, v.cfg, cfg)
		}
		t.Run(n, f)
	}
}
