//author: richard
package translate

import "github.com/advancevillage/3rd/https"

type Translate interface {
	Translate(q string, from, to string) (string, error)
}


type BaiDu struct {
	appId string
	key string
	request *https.Client
}
