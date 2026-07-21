package trie

import (
	"reflect"
	"strings"
	"testing"
)

func listToMap(items []string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, item := range items {
		m[item] = struct{}{}
	}
	return m
}

func newTestFilter(t *testing.T) Filter {
	t.Helper()
	f, err := NewFilter()
	if err != nil {
		t.Fatalf("new filter: %v", err)
	}
	return f
}

func TestLoad(t *testing.T) {
	filter := newTestFilter(t)
	if err := filter.Load(strings.NewReader("read")); err != nil {
		t.Errorf("fail to load dict %v", err)
	}
	if got := filter.FindAll("read"); len(got) == 0 {
		t.Errorf("load dict empty")
	}
}

func TestSensitiveFilter(t *testing.T) {
	filter := newTestFilter(t)
	filter.AddWord("有一个东西")
	filter.AddWord("一个东西")
	filter.AddWord("一个")
	filter.AddWord("东西")
	filter.AddWord("个东")

	testcases := []struct {
		Text   string
		Expect string
	}{
		{"我有一个东东西", "我有东"},
		{"我有一个东西", "我"},
		{"一个东西", ""},
		{"两个东西", "两西"},
		{"一个物体", "物体"},
	}

	for _, tc := range testcases {
		if got := filter.Filter(tc.Text); got != tc.Expect {
			t.Fatalf("filter %s, got %s, expect %s", tc.Text, got, tc.Expect)
		}
	}
}

func TestSensitiveValidateSingleword(t *testing.T) {
	filter := newTestFilter(t)
	filter.AddWord("东")

	testcases := []struct {
		Text        string
		ExpectPass  bool
		ExpectFirst string
	}{
		{"两个东西", false, "东"},
	}

	for _, tc := range testcases {
		if pass, first := filter.Validate(tc.Text); pass != tc.ExpectPass || first != tc.ExpectFirst {
			t.Fatalf("validate %s, got %v, %s, expect %v, %s", tc.Text, pass, first, tc.ExpectPass, tc.ExpectFirst)
		}
	}
}

func TestSensitiveValidate(t *testing.T) {
	filter := newTestFilter(t)
	filter.AddWord("有一个东西")
	filter.AddWord("一个东西")
	filter.AddWord("一个")
	filter.AddWord("东西")
	filter.AddWord("个东")
	filter.AddWord("有一个东西")
	filter.AddWord("一个东西")
	filter.AddWord("一个")
	filter.AddWord("东西")

	testcases := []struct {
		Text        string
		ExpectPass  bool
		ExpectFirst string
	}{
		{"我有一@ |个东东西", false, "一个"},
		{"我有一个东东西", false, "一个"},
		{"我有一个东西", false, "一个"},
		{"一个东西", false, "一个"},
		{"两个东西", false, "个东"},
		{"一样东西", false, "东西"},
	}

	for _, tc := range testcases {
		if pass, first := filter.Validate(tc.Text); pass != tc.ExpectPass || first != tc.ExpectFirst {
			t.Errorf("validate %s, got %v, %s, expect %v, %s", tc.Text, pass, first, tc.ExpectPass, tc.ExpectFirst)
		}
	}
}

func TestSensitiveReplace(t *testing.T) {
	filter := newTestFilter(t)
	filter.AddWord("有一个东西")
	filter.AddWord("一个东西")
	filter.AddWord("一个")
	filter.AddWord("东西")
	filter.AddWord("个东")

	testcases := []struct {
		Text   string
		Expect string
	}{
		{"我有一个东东西", "我有*****"},
		{"我有一个东西", "我*****"},
		{"一个东西", "****"},
		{"两个东西", "两***"},
		{"一个物体", "**物体"},
	}

	for _, tc := range testcases {
		if got := filter.Replace(tc.Text, '*'); got != tc.Expect {
			t.Fatalf("replace %s, got %s, expect %s", tc.Text, got, tc.Expect)
		}
	}
}

func TestSensitiveFindAll(t *testing.T) {
	filter := newTestFilter(t)
	filter.AddWord("有一个东西")
	filter.AddWord("一个东西")
	filter.AddWord("一个")
	filter.AddWord("东西")
	filter.AddWord("个东")

	testcases := []struct {
		Text   string
		Expect []string
	}{
		{"我有一个东东西", []string{"一个", "个东", "东西"}},
		{"我有一个东西", []string{"有一个东西", "一个", "一个东西", "个东", "东西"}},
		{"一个东西", []string{"一个", "一个东西", "个东", "东西"}},
		{"两个东西", []string{"个东", "东西"}},
		{"一个物体", []string{"一个"}},
	}

	for _, tc := range testcases {
		if got := filter.FindAll(tc.Text); !reflect.DeepEqual(listToMap(tc.Expect), listToMap(got)) {
			t.Errorf("findall %s, got %s, expect %s", tc.Text, got, tc.Expect)
		}
	}
}

func TestSensitiveFindallSingleword(t *testing.T) {
	filter := newTestFilter(t)
	filter.AddWord("东")

	testcases := []struct {
		Text   string
		Expect []string
	}{
		{"两个东西", []string{"东"}},
	}

	for _, tc := range testcases {
		if got := filter.FindAll(tc.Text); !reflect.DeepEqual(listToMap(tc.Expect), listToMap(got)) {
			t.Fatalf("findall %s, got %s, expect %s", tc.Text, got, tc.Expect)
		}
	}
}

func TestMatchFirstWithRemoveNoise(t *testing.T) {
	filter := newTestFilter(t)
	filter.AddWord("东西")

	if found, word := filter.MatchFirst("有东 西哈", WithRemoveNoise(true)); !found || word != "东西" {
		t.Fatalf("match with noise removal, got %v %s, expect true 东西", found, word)
	}
	if found, _ := filter.MatchFirst("有东 西哈"); found {
		t.Fatalf("match without noise removal should miss")
	}
}
