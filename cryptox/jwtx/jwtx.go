package jwtx

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/advancevillage/3rd/logx"
	"github.com/advancevillage/3rd/x"
	"github.com/golang-jwt/jwt/v5"
)

type JwtX interface {
	Sign(ctx context.Context, b x.Builder) (string, error)
	Valid(ctx context.Context, token string) bool
	Parse(ctx context.Context, token string) (x.Builder, error)
}

type JwtXOption interface {
	apply(*jwtxOption)
}

func WithSecretKey(sk string) JwtXOption {
	return newFuncJwtXOption(func(o *jwtxOption) {
		o.sk = sk
	})
}

func WithExpireTime(exp time.Duration) JwtXOption {
	return newFuncJwtXOption(func(o *jwtxOption) {
		o.exp = exp
	})
}

func WithAppName(app string) JwtXOption {
	return newFuncJwtXOption(func(o *jwtxOption) {
		o.iss = app
		o.sub = app
		o.aud = app
	})
}

type jwtxOption struct {
	sk  string
	iss string
	sub string
	aud string
	exp time.Duration
}

var defaultJwtXOptions = jwtxOption{
	sk:  "advancevillage",
	iss: "advancevillage",
	exp: time.Hour,
	sub: "advancevillage",
	aud: "advancevillage",
}

type funcJwtXOption struct {
	f func(*jwtxOption)
}

func (fdo *funcJwtXOption) apply(do *jwtxOption) {
	fdo.f(do)
}

func newFuncJwtXOption(f func(*jwtxOption)) *funcJwtXOption {
	return &funcJwtXOption{
		f: f,
	}
}

var _ JwtX = (*jwtx)(nil)

type jwtx struct {
	opts   jwtxOption
	logger logx.ILogger
}

func NewJwtXClient(ctx context.Context, logger logx.ILogger, opt ...JwtXOption) (JwtX, error) {
	// 1. 初始化配置
	opts := defaultJwtXOptions
	for _, o := range opt {
		o.apply(&opts)
	}

	// 2. 初始化对象
	j := &jwtx{opts: opts, logger: logger}

	return j, nil
}

func (j *jwtx) Sign(ctx context.Context, b x.Builder) (string, error) {
	// 1. 初始化参数
	var (
		q = url.Values{}
	)
	for k, v := range b.Build() {
		q.Add(k, fmt.Sprint(v))
	}
	// 2. 编码内容
	var (
		claims = jwtxCtx{
			Payload: q.Encode(), // 敏感信息
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    j.opts.iss,                                     // 签发者
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.opts.exp)), // 过期时间
				NotBefore: jwt.NewNumericDate(time.Now()),                 // 生效时间
				IssuedAt:  jwt.NewNumericDate(time.Now()),                 // 签发时间
				Subject:   j.opts.sub,                                     // 主题
				Audience:  jwt.ClaimStrings{j.opts.aud},                   // 受众
			},
		}
	)
	// 3. 签名
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	sign, err := t.SignedString([]byte(j.opts.sk))
	if err != nil {
		j.logger.Errorw(ctx, "jwt sign failed", "err", err, "context", q.Encode())
	}
	return sign, err
}

func (j *jwtx) Valid(ctx context.Context, token string) bool {
	t, err := j.parse(ctx, token)
	if err != nil {
		return false
	}
	return t.Valid
}

func (j *jwtx) Parse(ctx context.Context, token string) (x.Builder, error) {
	// 1. 解析
	t, err := j.parse(ctx, token)
	if err != nil {
		return nil, err
	}
	// 2. 解析内容
	claims, ok := t.Claims.(*jwtxCtx)
	if !ok || !t.Valid {
		return nil, JWTX_PARSE_ERROR
	}
	// 3. 提取内容
	q, err := url.ParseQuery(claims.Payload)
	if err != nil {
		return nil, err
	}
	// 4. 构建内容
	var opts []x.Option
	for k, v := range q {
		opts = append(opts, x.WithKV(k, v[0]))
	}
	return x.NewBuilder(opts...), nil
}

func (j *jwtx) parse(ctx context.Context, token string) (*jwt.Token, error) {
	// 1. 解析
	t, err := jwt.ParseWithClaims(token, &jwtxCtx{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.opts.sk), nil
	})
	if err != nil {
		j.logger.Errorw(ctx, "jwt parse failed", "err", err, "token", token)
		return nil, err
	}
	return t, nil
}
