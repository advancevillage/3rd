package dbx

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/kelindar/bitmap"
	"github.com/tencentyun/cos-go-sdk-v5"
)

type S3 interface {
	Uploader
	Operator
	Downloader
}

type Operator interface {
	Url(ctx context.Context, name string) (string, error)
	Exist(ctx context.Context, name string) (bool, error)
	Clean(cxx context.Context, name string) error
}

type Downloader interface {
	Download(ctx context.Context, name string) (string, error)
}

type Uploader interface {
	MultiUpload(ctx context.Context, name string, totalPart int) (MultiPartUploader, error)
	ResumeUpload(ctx context.Context, snapshot string) (MultiPartUploader, error)
}

var _ S3 = (*TxCos)(nil)

type TxCos struct {
	c  *cos.Client
	ak string
	sk string
}

// cos://ak:sk@bucket/region
func NewCosClient(ctx context.Context, dsn string) (S3, error) {
	ak, sk, bucket, region, err := ParseCosUrl(dsn)
	if err != nil {
		return nil, err
	}
	return NewCosS3(ctx, bucket, region, ak, sk)
}

func ParseCosUrl(dsn string) (ak, sk, bkt, rgn string, err error) {
	u, err := url.Parse(dsn)
	if err != nil {
		return ak, sk, bkt, rgn, err
	}
	if u.Scheme != "cos" {
		return ak, sk, bkt, rgn, errors.New("cos: invalid scheme")
	}
	var ok bool
	ak = u.User.Username()
	sk, ok = u.User.Password()
	bkt, rgn = u.Host, strings.TrimPrefix(u.Path, "/")
	if !ok {
		return ak, sk, bkt, rgn, errors.New("cos: invalid sk")
	}
	return ak, sk, bkt, rgn, nil
}

func NewCosS3(ctx context.Context, bucket string, region string, ak string, sk string) (S3, error) {
	b, err := cos.NewBucketURL(bucket, region, true)
	if err != nil {
		return nil, err
	}
	c := cos.NewClient(&cos.BaseURL{BucketURL: b}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  ak,
			SecretKey: sk,
		},
	})
	return &TxCos{c: c, ak: ak, sk: sk}, nil
}

func (t *TxCos) Download(ctx context.Context, name string) (string, error) {
	return t.getPresignedUrl(ctx, http.MethodGet, name)
}

func (t *TxCos) Url(ctx context.Context, name string) (string, error) {
	return t.getPresignedUrl(ctx, http.MethodGet, name)
}

func (t *TxCos) Exist(ctx context.Context, name string) (bool, error) {
	_, err := t.c.Object.Head(ctx, name, nil)
	switch {
	case err == nil:
		return true, nil

	case cos.IsNotFoundError(err):
		return false, nil

	default:
		return false, err
	}
}

func (t *TxCos) Clean(ctx context.Context, name string) error {
	_, err := t.c.Object.Delete(ctx, name, nil)
	switch {
	case err == nil:
		return nil

	case cos.IsNotFoundError(err):
		return nil

	default:
		return err
	}
}

func (t *TxCos) MultiUpload(ctx context.Context, name string, totalPart int) (MultiPartUploader, error) {
	return newMultipartUploader(ctx, t.c, name, totalPart)
}

func (t *TxCos) ResumeUpload(ctx context.Context, snapshot string) (MultiPartUploader, error) {
	return newMultipartResumer(ctx, t.c, snapshot)
}

func (t *TxCos) getPresignedUrl(ctx context.Context, httpMethod string, name string) (string, error) {
	url, err := t.c.Object.GetPresignedURL(ctx, httpMethod, name, t.ak, t.sk, time.Hour, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

type MultiPartUploader interface {
	// 上传ID
	Id(ctx context.Context) string
	// 分组大小1M~5G
	Write(ctx context.Context, partNumber int, body []byte) error
	// 上传进度
	Progress() float64
	// 上传快照
	String() string
}

var _ MultiPartUploader = (*multiprtUploader)(nil)

type multiprtUploader struct {
	c         *cos.Client
	name      string
	parts     *cos.CompleteMultipartUploadOptions
	uploadId  string
	totalPart int
	bits      bitmap.Bitmap
}

func newMultipartResumer(ctx context.Context, c *cos.Client, dsn string) (*multiprtUploader, error) {
	// 1. 解析协议
	u, err := url.ParseQuery(dsn)
	if err != nil {
		return nil, err
	}
	// 2. 获取上传id
	var (
		id   = u.Get("id")
		name = u.Get("name")
	)
	total, err := strconv.Atoi(u.Get("total"))
	if err != nil {
		return nil, err
	}

	// 3. 初始化
	mp := &multiprtUploader{c: c, name: name, totalPart: total, parts: &cos.CompleteMultipartUploadOptions{}, uploadId: id}

	// 4. 已上传分片
	for k := range u {
		if !strings.HasPrefix(k, "part.") {
			continue
		}
		part, err := strconv.Atoi(k[5:])
		if err != nil {
			return nil, err
		}
		if part < 1 || part > total {
			continue
		}
		mp.parts.Parts = append(mp.parts.Parts, cos.Object{
			PartNumber: part,
			ETag:       u.Get(k),
		})
		mp.bits.Set(uint32(part - 1))
	}
	return mp, nil
}

func newMultipartUploader(ctx context.Context, c *cos.Client, name string, totalPart int) (*multiprtUploader, error) {
	var (
		mp  = &multiprtUploader{c: c, name: name, totalPart: totalPart, parts: &cos.CompleteMultipartUploadOptions{}}
		err error
	)
	mp.uploadId, err = mp.initiate(ctx, mp.name)
	if err != nil {
		return nil, err
	}
	return mp, nil
}

// PartNmber: 从0开始
func (mp *multiprtUploader) Write(ctx context.Context, partNumber int, body []byte) error {
	r := bytes.NewBuffer(body)
	partNumber = partNumber % mp.totalPart
	etag, err := mp.uploadPart(ctx, mp.name, mp.uploadId, partNumber+1, r)
	if err != nil {
		return err
	}
	mp.parts.Parts = append(mp.parts.Parts, cos.Object{
		PartNumber: partNumber + 1, // COS从1开始
		ETag:       etag,
	})
	mp.bits.Set(uint32(partNumber))
	if mp.bits.Count() >= mp.totalPart {
		return mp.complete(ctx, mp.name, mp.uploadId, mp.parts)
	}
	return nil
}

func (mp *multiprtUploader) Id(ctx context.Context) string {
	return mp.uploadId
}

func (mp *multiprtUploader) initiate(ctx context.Context, name string) (string, error) {
	r, reply, err := mp.c.Object.InitiateMultipartUpload(ctx, name, nil)
	if err != nil {
		return "", err
	}
	if reply.StatusCode != http.StatusOK {
		return "", errors.New("cos: initiate multipart upload failed")
	}
	return r.UploadID, nil
}

func (mp *multiprtUploader) uploadPart(ctx context.Context, name string, uploadId string, partNumber int, r io.Reader) (string, error) {
	reply, err := mp.c.Object.UploadPart(ctx, name, uploadId, partNumber, r, nil)
	if err != nil {
		return "", err
	}
	return reply.Header.Get("Etag"), nil
}

func (mp *multiprtUploader) complete(ctx context.Context, name, uploadId string, opt *cos.CompleteMultipartUploadOptions) error {
	sort.SliceStable(opt.Parts, func(i, j int) bool {
		return opt.Parts[i].PartNumber < opt.Parts[j].PartNumber
	})
	_, reply, err := mp.c.Object.CompleteMultipartUpload(ctx, name, uploadId, opt)
	if err != nil {
		return err
	}
	if reply.StatusCode != http.StatusOK {
		return errors.New("cos: complete multipart upload failed")
	}
	return nil
}

func (mp *multiprtUploader) String() string {
	u := url.Values{}
	u.Add("id", mp.uploadId)
	u.Add("name", mp.name)
	u.Add("total", strconv.Itoa(mp.totalPart))
	u.Add("progress", strconv.FormatFloat(mp.Progress(), 'f', 2, 64))
	for i := range mp.parts.Parts {
		u.Add(fmt.Sprintf("part.%d", mp.parts.Parts[i].PartNumber), mp.parts.Parts[i].ETag)
	}
	return u.Encode()
}

func (mp *multiprtUploader) Progress() float64 {
	return float64(mp.bits.Count()) / float64(mp.totalPart)
}
