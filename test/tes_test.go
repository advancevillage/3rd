//author: richard
package test

import (
	"encoding/json"
	"github.com/advancevillage/3rd/logs"
	"github.com/advancevillage/3rd/storages"
	"testing"
)

func TestNewTES(t *testing.T) {
	logger, err := logs.NewTxtLogger("tes.log", 128, 3)
	if err != nil {
		t.Error(err.Error())
		return
	}
	urls := []string{"http://localhost:9200"}
	tes, err := storages.NewTES(urls, logger)
	if err != nil {
		t.Error(err.Error())
		return
	}
	goods := struct {
		GoodName  string	`json:"good_name"`
		GoodPrice float32	`json:"good_price"`
		GoodStock int		`json:"good_stock"`
	}{
		GoodName: "cars",
		GoodPrice: 99,
		GoodStock: 102,
	}
	index := "kelly"
	id := "richard"
	err = tes.CreateDocument(index, id, goods)
	if err != nil {
		t.Error(err.Error())
		return
	}
	err = tes.UpdateDocument(index, id, map[string]interface{}{"good_price": 106,})
	if err != nil {
		t.Error(err.Error())
		return
	}
}

func TestTES_CreateStorage(t *testing.T) {
	logger, err := logs.NewTxtLogger("tes.log", 128, 3)
	if err != nil {
		t.Error(err.Error())
		return
	}
	urls := []string{"http://localhost:9200"}
	tes, err := storages.NewTES(urls, logger)
	if err != nil {
		t.Error(err.Error())
		return
	}
	goods := struct {
		GoodName  string	`json:"good_name"`
		GoodPrice float32	`json:"good_price"`
		GoodStock int		`json:"good_stock"`
	}{
		GoodName: "cars",
		GoodPrice: 99,
		GoodStock: 102,
	}
	buf, err := json.Marshal(goods)
	if err != nil {
		t.Error(err.Error())
		return
	}
	key := "kelly"
	err = tes.CreateStorage(key, buf)
	if err != nil {
		t.Error(err.Error())
		return
	}
	ret, err := tes.QueryDocument(storages.ESDefaultIndex, key)
	if err != nil {
		t.Error(err.Error())
		return
	}
	t.Log(ret)
}
