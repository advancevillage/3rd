package dbx

import (
	"fmt"
	"os"
	"testing"

	"github.com/advancevillage/3rd/mathx"
	"github.com/stretchr/testify/assert"
)

var ldbTestData = map[string]struct {
	key    string
	prefix string
	value  string
}{
	"case1": {
		key:    mathx.RandStr(8),
		value:  mathx.RandStr(16),
		prefix: "t-",
	},
	"case2": {
		key:    mathx.RandStr(8),
		value:  mathx.RandStr(16),
		prefix: "t-",
	},
	"case3": {
		key:    mathx.RandStr(8),
		value:  mathx.RandStr(16),
		prefix: "t-",
	},
	"case4": {
		key:    mathx.RandStr(8),
		value:  mathx.RandStr(16),
		prefix: "t-",
	},
	"case5": {
		key:    mathx.RandStr(8),
		value:  mathx.RandStr(16),
		prefix: "t-",
	},
	"case6": {
		key:    mathx.RandStr(8),
		value:  mathx.RandStr(16),
		prefix: "t-",
	},
	"case7": {
		key:    mathx.RandStr(8),
		value:  mathx.RandStr(16),
		prefix: "t-",
	},
	"case8": {
		key:    mathx.RandStr(8),
		value:  mathx.RandStr(16),
		prefix: "t-",
	},
	"case9": {
		key:    mathx.RandStr(8),
		value:  mathx.RandStr(16),
		prefix: "t-",
	},
	"case10": {
		key:    mathx.RandStr(8),
		value:  mathx.RandStr(16),
		prefix: "t-",
	},
}

func Test_mem_ldb(t *testing.T) {
	var s, err = NewMemoryStore()
	if err != nil {
		t.Fatal(err)
		return
	}
	for n, p := range ldbTestData {
		f := func(t *testing.T) {
			var tk = fmt.Sprintf("%s%s", p.prefix, p.key)
			err = s.Put([]byte(tk), []byte(p.value))
			if err != nil {
				t.Fatal(err)
				return
			}
			a, err := s.Get([]byte(tk))
			if err != nil {
				t.Fatal(err)
				return
			}
			assert.Equal(t, a, []byte(p.value))
			kv, err := s.Range([]byte(p.prefix))
			if err != nil {
				t.Fatal(err)
				return
			}
			assert.Equal(t, kv[tk], p.value)
			err = s.Del([]byte(tk))
			if err != nil {
				t.Fatal(err)
				return
			}
		}
		t.Run(n, f)
	}
}

func Test_p_ldb(t *testing.T) {
	var pdir, _ = os.Getwd()
	var dir = fmt.Sprintf("%s/%s", pdir, "temp")
	os.MkdirAll(dir, os.ModePerm)
	var s, err = NewPersistentStore(dir)
	if err != nil {
		t.Fatal(err)
		return
	}
	for n, p := range ldbTestData {
		f := func(t *testing.T) {
			var tk = fmt.Sprintf("%s%s", p.prefix, p.key)
			err = s.Put([]byte(tk), []byte(p.value))
			if err != nil {
				t.Fatal(err)
				return
			}
			a, err := s.Get([]byte(tk))
			if err != nil {
				t.Fatal(err)
				return
			}
			assert.Equal(t, a, []byte(p.value))
			kv, err := s.Range([]byte(p.prefix))
			if err != nil {
				t.Fatal(err)
				return
			}
			assert.Equal(t, kv[tk], p.value)
			err = s.Del([]byte(tk))
			if err != nil {
				t.Fatal(err)
				return
			}
		}
		t.Run(n, f)
	}
	os.Remove(dir)
}
