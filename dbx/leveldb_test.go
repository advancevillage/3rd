package dbx

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var ldbTestData = map[string]struct {
	key    string
	prefix string
	value  string
}{
	"case1": {
		key:    randsString(8),
		value:  randsString(16),
		prefix: "t-",
	},
	"case2": {
		key:    randsString(8),
		value:  randsString(16),
		prefix: "t-",
	},
	"case3": {
		key:    randsString(8),
		value:  randsString(16),
		prefix: "t-",
	},
	"case4": {
		key:    randsString(8),
		value:  randsString(16),
		prefix: "t-",
	},
	"case5": {
		key:    randsString(8),
		value:  randsString(16),
		prefix: "t-",
	},
	"case6": {
		key:    randsString(8),
		value:  randsString(16),
		prefix: "t-",
	},
	"case7": {
		key:    randsString(8),
		value:  randsString(16),
		prefix: "t-",
	},
	"case8": {
		key:    randsString(8),
		value:  randsString(16),
		prefix: "t-",
	},
	"case9": {
		key:    randsString(8),
		value:  randsString(16),
		prefix: "t-",
	},
	"case10": {
		key:    randsString(8),
		value:  randsString(16),
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

func randsString(n int) string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	r := rand.New(rand.NewSource(99))
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[r.Intn(len(letterRunes))]
	}
	return string(b)
}
