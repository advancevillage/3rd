//author: richard
package test

import (
	"encoding/json"
	"github.com/advancevillage/3rd/logs"
	"github.com/advancevillage/3rd/pools"
	"github.com/advancevillage/3rd/queues"
	"github.com/advancevillage/3rd/utils"
	"testing"
	"time"
)

func TestTQueue(t *testing.T) {
	queue := queues.NewTQueue(10)
	e := &queues.Element{}
	logger, err := logs.NewTxtLogger("t.log", 128, 4)
	if err != nil {
		t.Error(err.Error())
	}
	pool := pools.NewGoPool(4, logger)
	handler := func(e interface{}) {
		ele := queues.Element{}
		buf, err := json.Marshal(e)
		if err != nil {
			t.Error(err.Error())
		} else {
			err = json.Unmarshal(buf, &ele)
			if err != nil {
				t.Error(err.Error())
			} else {
				logger.Info(string(ele.Body))
			}
		}
	}
	go func() {
		for {
			if queue.Full() {
				continue
			}
			e.Key = utils.RandsString(10)
			e.Body = []byte(utils.UUID())
			err = queue.Enqueue(e)
			if err != nil {
				t.Error(err.Error())
			}
			time.Sleep(time.Second)
		}
	}()
	for  {
		if queue.Empty() {
			continue
		}
		ele, err := queue.Dequeue()
		if err != nil {
			t.Error(err.Error())
		}
		err = pool.Process(ele, handler)
		if err != nil {
			t.Error(err.Error())
		}
		time.Sleep(time.Second)/**/
	}
}
