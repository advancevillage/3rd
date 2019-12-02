//author: richard
//@note: tes = tradition elastic search
package storages

import (
	"3rd/logs"
	"context"
	"github.com/olivere/elastic/v7"
)

//@link: https://elasticsearch.cn/article/6178
//@link: https://www.kancloud.cn/liupengjie/go/570150
//@link: https://www.do1618.com/archives/1355/no-elasticsearch-node-available/

func NewTES(urls []string, logger logs.Logs) (*TES, error) {
	var err error
	tes := &TES{}
	tes.urls = urls
	tes.logger = logger
	tes.conn, err = elastic.NewClient(elastic.SetSniff(false), elastic.SetURL(tes.urls...))
	if err != nil {
		tes.logger.Error(err.Error())
		return nil, err
	}
	for i :=range urls {
		info, code, err := tes.conn.Ping(urls[i]).Do(context.Background())
		if err != nil {
			tes.logger.Error(err.Error())
			return nil, err
		}
		tes.logger.Info("info=%v, code=%d", info, code)
	}
	return tes, nil
}