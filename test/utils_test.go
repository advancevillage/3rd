//author: richard
package test

import (
	"3rd/times"
	"3rd/utils"
	"testing"
)

func TestRandsString(t *testing.T) {
	for i := 0; i < 5000; i++ {
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

func TestRandInt(t *testing.T) {
	for i := 0; i < 100; i++ {
		t.Log(utils.RandsInt(100))
	}
}

func TestTime(t *testing.T) {
	for i := 0; i < 500; i++ {
		t.Log(times.Timestamp())
		t.Log(times.TimeString())
	}
}

func BenchmarkTime(b *testing.B) {
	for i := 0; i < b.N; i++ {
		b.Log(times.Timestamp())
		b.Log(times.TimeString())
	}
}

func BenchmarkUUID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		str := utils.UUID()
		b.Log(str)
	}
}
