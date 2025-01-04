package dbx_test

import (
	"context"
	"testing"

	"github.com/advancevillage/3rd/dbx"
	"github.com/advancevillage/3rd/mathx"
	"golang.org/x/sync/errgroup"
)

func Test_S3(t *testing.T) {
	var data = map[string]struct {
		ak     string
		sk     string
		bucket string
		region string
		sence  string
		ext    map[string]interface{}
	}{
		"case1": {
			bucket: "xmagic-1259635961",
			region: "ap-shanghai",
			ak:     "AKIDmPxIQO9h1B8ECzax74pOlupXmodEdrsg",
			sk:     "e4KOFsHTRBk6FBuQjpiOchnkZEm7HEyV",
			sence:  "upload",
			ext: map[string]interface{}{
				"name":  "test/test.txt",
				"total": 12,
				"parts": map[int]string{
					0:  mathx.RandStr(1 << 20),
					1:  mathx.RandStr(1 << 20),
					2:  mathx.RandStr(1 << 20),
					3:  mathx.RandStr(1 << 20),
					4:  mathx.RandStr(1 << 20),
					5:  mathx.RandStr(1 << 20),
					6:  mathx.RandStr(1 << 20),
					7:  mathx.RandStr(1 << 20),
					8:  mathx.RandStr(1 << 20),
					9:  mathx.RandStr(1 << 20),
					10: mathx.RandStr(1 << 20),
					11: mathx.RandStr(1 << 11),
				},
			},
		},
	}

	for n, v := range data {
		f := func(t *testing.T) {
			c, err := dbx.NewCosS3(context.TODO(), v.bucket, v.region, v.ak, v.sk)
			if err != nil {
				t.Fatal(err)
				return
			}
			switch v.sence {
			case "upload":
				name, ok := v.ext["name"].(string)
				if !ok {
					t.Fatal("param name is not valid")
					return
				}
				total, ok := v.ext["total"].(int)
				if !ok || total <= 0 {
					t.Fatal("param total is not valid")
					return
				}
				parts, ok := v.ext["parts"].(map[int]string)
				if err != nil {
					t.Fatal("param parts is not valid")
					return
				}
				up, err := c.MultiUpload(context.TODO(), name, total)
				if err != nil {
					t.Fatal(err)
					return
				}
				t.Logf("case=%s, uoloadId:%s", n, up.Id(context.TODO()))

				var (
					ch = make(chan struct{}, 3)
					g  = new(errgroup.Group)
				)
				for i, v := range parts {
					partNumber := i
					body := []byte(v)
					ch <- struct{}{}
					g.Go(func() error {
						defer func() {
							<-ch
						}()
						return up.Write(context.TODO(), partNumber, body)
					})
				}
				if err := g.Wait(); err != nil {
					t.Fatal(err)
					return
				}
			case "download":

			case "exist":

			}
		}
		t.Run(n+"_"+v.sence, f)
	}
}
