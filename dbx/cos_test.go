package dbx_test

import (
	"context"
	"os"
	"testing"

	"github.com/advancevillage/3rd/dbx"
	"github.com/advancevillage/3rd/mathx"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func Test_S3(t *testing.T) {
	var (
		ak = os.Getenv("COS_AK")
		sk = os.Getenv("COS_SK")
	)
	var data = map[string]struct {
		bucket string
		region string
		ext    map[string]any
	}{
		"case1": {
			bucket: "xmagic-1259635961",
			region: "ap-shanghai",
			ext: map[string]any{
				"name":  "test/" + mathx.RandStr(5) + ".txt",
				"total": 6,
				"parts": map[int]string{
					0: mathx.RandStr(1 << 20),
					1: mathx.RandStr(1 << 20),
					2: mathx.RandStr(1 << 20),
					3: mathx.RandStr(1 << 20),
					4: mathx.RandStr(1 << 20),
					5: mathx.RandStr(1 << 11),
				},
			},
		},
		"case2": {
			bucket: "xmagic-1259635961",
			region: "ap-shanghai",
			ext: map[string]any{
				"name":  "test/" + mathx.RandStr(5) + ".txt",
				"total": 1,
				"parts": map[int]string{
					0: mathx.RandStr(1 << 11),
				},
			},
		},
	}

	for n, v := range data {
		f := func(t *testing.T) {
			c, err := dbx.NewCosS3(context.TODO(), v.bucket, v.region, ak, sk)
			if err != nil {
				t.Fatal(err)
				return
			}
			// 上传文件
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
			if !ok {
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

			// 存在性检查
			exist, err := c.Exist(context.TODO(), name)
			if err != nil {
				t.Fatal(err)
				return
			}
			assert.Equal(t, true, exist)

			// 下载文件检查
			uri, err := c.Download(context.TODO(), name)
			if err != nil {
				t.Fatal(err)
				return
			}
			t.Logf("case=%s, download uri:%s", n, uri)

			// 删除文件
			err = c.Clean(context.TODO(), name)
			if err != nil {
				t.Fatal(err)
				return
			}

			// 存在性检查
			exist, err = c.Exist(context.TODO(), name)
			if err != nil {
				t.Fatal(err)
				return
			}
			assert.Equal(t, false, exist)
		}
		t.Run(n, f)
	}
}

func Test_ParseCosUrl(t *testing.T) {
	var data = map[string]struct {
		dsn string
		ak  string
		sk  string
		bkt string
		rgn string
	}{
		"case1": {
			dsn: "cos://1122:3344@xmagic-1259635961/ap-shanghai",
			ak:  "1122",
			sk:  "3344",
			bkt: "xmagic-1259635961",
			rgn: "ap-shanghai",
		},
	}
	for n, v := range data {
		f := func(t *testing.T) {
			ak, sk, bkt, rgn, err := dbx.ParseCosUrl(v.dsn)
			assert.Nil(t, err)
			assert.Equal(t, v.ak, ak)
			assert.Equal(t, v.sk, sk)
			assert.Equal(t, v.bkt, bkt)
			assert.Equal(t, v.rgn, rgn)
		}
		t.Run(n, f)
	}
}
