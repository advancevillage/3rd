//author: richard
package translate

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/advancevillage/3rd/https"
	"github.com/advancevillage/3rd/times"
)

func NewBaiDuTranslate(appId string, key string) *BaiDu {
	return &BaiDu{
		appId: appId,
		key: key,
		request: https.NewRequest(nil, 60, 2),
	}
}

//@doc: https://api.fanyi.baidu.com/doc/11
func (s *BaiDu) Translate(q string, from, to string) (string, error) {
	salt := times.Timestamp()
	sign := fmt.Sprintf("%s%s%d%s", s.appId, q, salt, s.key)
	sign = fmt.Sprintf("%x", md5.Sum([]byte(sign)))
	uri := fmt.Sprintf("http://api.fanyi.baidu.com/api/trans/vip/translate?q=%s&appid=%s&salt=%d&from=%s&to=%s&sign=%s",
		q, s.appId, salt, from, to, sign)
	buf, err := s.request.GET(uri, nil, nil)
	if err != nil {
		return "", err
	}
	body := struct {
		From string `json:"from"`
		To   string `json:"to"`
		Result []struct{
			Src string `json:"src"`
			Dst string `json:"dst"`
		} `json:"trans_result"`
	}{}
	err = json.Unmarshal(buf, &body)
	if err != nil {
		return "", err
	}
	if len(body.Result) < 1 {
		return "", errors.New("translate no result")
	}
	return body.Result[0].Dst, nil
}