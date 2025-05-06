package dbx_test

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/advancevillage/3rd/dbx"
	"github.com/advancevillage/3rd/logx"
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
		"case3": {
			bucket: "xmagic-1259635961",
			region: "accelerate",
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
			name = strings.ToUpper(name)
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
		"case2": {
			dsn: "cos://1122:3344@xmagic-1259635961/accelerate",
			ak:  "1122",
			sk:  "3344",
			bkt: "xmagic-1259635961",
			rgn: "accelerate",
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

func Test_S3_Resumer(t *testing.T) {
	logger, err := logx.NewLogger("debug")
	assert.Nil(t, err)
	ctx := context.WithValue(context.TODO(), logx.TraceId, mathx.UUID())
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
			region: "accelerate",
			ext: map[string]any{
				"name":  "TEST/" + mathx.RandStr(5) + ".txt",
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
	}
	for n, v := range data {
		f := func(t *testing.T) {
			name, ok := v.ext["name"].(string)
			assert.Equal(t, true, ok)
			total, ok := v.ext["total"].(int)
			assert.Equal(t, true, ok)
			assert.Equal(t, true, total > 0)
			parts, ok := v.ext["parts"].(map[int]string)
			assert.Equal(t, true, ok)

			// 上传文件
			up, err := newS3Locker(ctx, logger, v.bucket, v.region, ak, sk)
			assert.Nil(t, err)

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
					return up.Upload(ctx, name, total, partNumber, body)
				})
			}
			if err := g.Wait(); err != nil {
				t.Fatal(err)
				return
			}

			c, err := dbx.NewCosS3(ctx, v.bucket, v.region, ak, sk)
			assert.Nil(t, err)

			// 存在性检查
			exist, err := c.Exist(ctx, name)
			assert.Nil(t, err)
			assert.Equal(t, true, exist)

			// 下载文件检查
			uri, err := c.Download(ctx, name)
			assert.Nil(t, err)
			t.Logf("case=%s, download uri:%s", n, uri)

			// 删除文件
			err = c.Clean(ctx, name)
			assert.Nil(t, err)

			// 存在性检查
			exist, err = c.Exist(ctx, name)
			assert.Nil(t, err)
			assert.Equal(t, false, exist)
		}
		t.Run(n, f)
	}
}

type s3locker struct {
	s3     dbx.S3
	cacher dbx.Cacher
	locker dbx.CacheLocker
	logger logx.ILogger
	preifx string
}

func newS3Locker(ctx context.Context, logger logx.ILogger, bkt string, rgn string, ak string, sk string) (*s3locker, error) {
	s3, err := dbx.NewCosS3(ctx, bkt, rgn, ak, sk)
	if err != nil {
		return nil, err
	}
	locker, err := dbx.NewCacheRedisLocker(ctx, logger)
	if err != nil {
		return nil, err
	}
	cacher, err := dbx.NewCacheRedis(ctx, logger)
	if err != nil {
		return nil, err
	}
	return &s3locker{
		s3:     s3,
		preifx: "lock:",
		cacher: cacher,
		locker: locker,
		logger: logger,
	}, nil
}

func (c *s3locker) Upload(ctx context.Context, name string, total int, index int, part []byte) error {
	// 1. 锁标识
	val := mathx.RandStr(5)

	// 2. 尝试加锁
	err := c.tryLock(ctx, c.preifx+name, val)
	if err != nil {
		return err
	}
	defer c.locker.Unlock(ctx, c.preifx+name, val)

	// 3. 上传文件
	cache := c.cacher.CreateStringCacher(ctx, "upload:"+name, time.Minute*10)

	snapshot, err := cache.Get(ctx)
	if err != nil {
		return err
	}
	var mp dbx.MultiPartUploader

	switch {
	// 非首分片上传
	case len(snapshot) > 0:
		mp, err = c.s3.ResumeUpload(ctx, snapshot)
	// 首分片上传
	default:
		mp, err = c.s3.MultiUpload(ctx, name, total)
	}
	if err != nil {
		return err
	}
	err = mp.Write(ctx, index, part)
	if err != nil {
		return err
	}
	snapshot = mp.String()
	err = cache.Set(ctx, snapshot)
	if err != nil {
		return err
	}
	c.logger.Infow(ctx, "upload part success", "name", name, "index", index, "progress", mp.Progress())
	return nil
}

func (c *s3locker) tryLock(ctx context.Context, key string, val string) error {
	t := time.NewTicker(time.Millisecond * 200)
	defer t.Stop()
	n := 0

	for {
		lock, err := c.locker.Lock(ctx, key, val, time.Second*10)
		if err != nil {
			return err
		}
		if lock {
			return nil
		}
		if n > 100 {
			return errors.New("upload part timeout")
		}
		<-t.C
	}
}
