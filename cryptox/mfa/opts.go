package mfa

type MFAOption interface {
	apply(*mfaOption)
}

func WithEmail(email string) MFAOption {
	return newFuncMFAOption(func(o *mfaOption) {
		o.email = email
	})
}

func WithSecret(sk string) MFAOption {
	return newFuncMFAOption(func(o *mfaOption) {
		o.sk = sk
	})
}

type mfaOption struct {
	sk     string // 密钥
	email  string // 邮箱
	period uint   // 有效期
}

var defaultMFAOptions = mfaOption{
	sk:     "123456",
	email:  "richard.sun@softpart.run",
	period: 30,
}

type funcMFAOption struct {
	f func(*mfaOption)
}

func (fdo *funcMFAOption) apply(do *mfaOption) {
	fdo.f(do)
}

func newFuncMFAOption(f func(*mfaOption)) *funcMFAOption {
	return &funcMFAOption{
		f: f,
	}
}
