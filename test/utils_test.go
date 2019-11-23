//author: richard
package test

import (
	"3rd/utils"
	"testing"
)

func TestRandsString(t *testing.T) {
	for i := 0; i < 500; i++ {
		str := utils.RandsString(50)
		t.Log(str)
	}
}

func TestUUID(t *testing.T) {
	for i := 0; i < 10000; i++ {
		str := utils.UUID()
		t.Log(str)
	}
}
