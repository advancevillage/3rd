// Copyright 2014 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// 来源：https://github.com/google/btree/blob/479b5e81b0a93ec038d201b0b33d17db599531d3/btree.go
// 本地修改：拆分文件，适配 3rd 接口和 Option 构造；保留原节点算法。
package btree

import (
	"fmt"
	"reflect"

	"github.com/advancevillage/3rd/x"
)

// Option 配置 BTree，依次应用，后者覆盖前者。
type Option = x.Options[option]
type option struct {
	degree   int
	freelist *FreeList
}

// WithDegree 设置最小度数，必须至少为 2。默认 32。
func WithDegree(degree int) Option { return x.NewFuncOptions(func(o *option) { o.degree = degree }) }

// WithFreeList 指定共享节点缓存；不能为 nil，使用后不得复制 FreeList。
func WithFreeList(f *FreeList) Option { return x.NewFuncOptions(func(o *option) { o.freelist = f }) }

// NewBTree 创建空树，配置无效时返回错误。默认使用独立的节点缓存。
func NewBTree(opts ...Option) (BTree, error) {
	f, err := NewFreeList(DefaultFreeListSize)
	if err != nil {
		return nil, err
	}
	o := option{degree: 32, freelist: f}
	for _, opt := range opts {
		if opt == nil {
			return nil, fmt.Errorf("btree: nil option")
		}
		v := reflect.ValueOf(opt)
		switch v.Kind() {
		case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
			if v.IsNil() {
				return nil, fmt.Errorf("btree: nil option")
			}
		}
		opt.Apply(&o)
	}
	if o.degree < 2 || o.degree > int(^uint(0)>>1)/2 {
		return nil, fmt.Errorf("btree: invalid degree: %d", o.degree)
	}
	if o.freelist == nil {
		return nil, fmt.Errorf("btree: nil free list")
	}
	return &btree{degree: o.degree, cow: &copyOnWriteContext{freelist: o.freelist}}, nil
}
