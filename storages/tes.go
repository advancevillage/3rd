//author: richard
//@note: tes = tradition elastic search
package storages

import (
	"3rd/logs"
	"context"
	"encoding/json"
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
	tes.index  = "20170729"
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

//实现接口
func (tes *TES) CreateStorage(key string, body []byte) error {
	var object = make(map[string]interface{})
	var err error
	err = json.Unmarshal(body, &object)
	if err != nil {
		tes.logger.Error(err.Error())
		return err
	}
	return tes.CreateDocument(tes.index, key, object)
}

func (tes *TES) UpdateStorage(key string, body []byte) error {
	var object = make(map[string]interface{})
	var err error
	err = json.Unmarshal(body, &object)
	if err != nil {
		tes.logger.Error(err.Error())
		return err
	}
	return tes.CreateDocument(tes.index, key, object)
}

func (tes *TES) DeleteStorage(key ...string) error {
	for i := range key {
		err := tes.DeleteDocument(tes.index, key[i])
		if err != nil {
			tes.logger.Error(err.Error())
		} else {
			continue
		}
	}
	return nil
}

func (tes *TES) QueryStorage(key string) ([]byte, error) {
	//TODO
	return nil, nil
}

//创建一个文档,如果文档不存在则创建。如果存在则更新值
func (tes *TES) CreateDocument(index string, id string, body interface{}) error {
	_, err := tes.conn.Index().Index(index).Id(id).BodyJson(body).Do(context.Background())
	if err != nil {
		tes.logger.Error(err.Error())
		return err
	}
	return nil
}

func (tes *TES) DeleteDocument(index string, id string) error {
	_, err := tes.conn.Delete().Index(index).Id(id).Do(context.Background())
	if err != nil {
		tes.logger.Error(err.Error())
		return err
	}
	return nil
}

func (tes *TES) UpdateDocument(index string, id string, fields map[string]interface{}) error {
	_, err := tes.conn.Update().Index(index).Id(id).Doc(fields).Do(context.Background())
	if err != nil {
		tes.logger.Error(err.Error())
		return err
	}
	return nil
}