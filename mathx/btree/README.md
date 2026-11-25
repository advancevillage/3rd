# B-tree

内存有序集合，支持增删查、升降序及范围遍历、写时复制克隆和节点缓存。

来源：[google/btree](https://github.com/google/btree)，基线提交
[`479b5e81b0a93ec038d201b0b33d17db599531d3`](https://github.com/google/btree/tree/479b5e81b0a93ec038d201b0b33d17db599531d3)。
保留原始版权、Apache-2.0 LICENSE、设计说明和参考引用；本地修改主要为文件拆分、接口与 Option 构造，以及测试适配。

```go
import "github.com/advancevillage/3rd/mathx/btree"

func example() error {
    tree, err := btree.NewBTree(btree.WithDegree(32))
    if err != nil {
        return err
    }
    tree.ReplaceOrInsert(btree.Int(10))
    tree.Ascend(func(item btree.Item) bool {
        // 使用 item；返回 false 可以提前停止。
        return true
    })
    return nil
}
```

## 与来源 API 的差异

- `BTree` 由导出结构体变为接口，具体实现私有。
- `New(degree)` 改为 `NewBTree(WithDegree(degree))`，返回 `(BTree, error)`。
- `NewWithFreeList(degree, f)` 改为 `NewBTree(WithDegree(degree), WithFreeList(f))`。
- `NewFreeList(size)` 返回 `(*FreeList, error)`；负数容量返回错误，零容量关闭节点缓存。
- 默认 degree 为 32，节点缓存容量为 32；默认各树缓存独立。可显式传入 FreeList 共享缓存。
- 无效 degree、nil Option、typed-nil Option 和显式 nil FreeList 返回错误。选项按顺序覆盖，最终配置统一校验。
- `Clone()` 返回 `BTree`，其他树操作保持来源语义；插入 nil 仍会 panic。

## 使用约束

Item.Less 必须满足严格弱序，双方互不小于视为同一键，插入时替换旧值。
`AscendRange` 使用 `[lower, upper)`，`DescendRange` 使用 `(lower, upper]`。
同一棵树仅在没有写操作时支持并发读；读写或多写并发需外部同步。
Clone 不可与源树操作并发；完成后源树和克隆树可分别使用。
Clone 不深复制 Item，调用方不得修改已入树元素的排序字段；共享可变 Item 需自行同步。
FreeList 带锁，使用后不得复制。零值树不可用，应通过 NewBTree 构造。

原独立内存对比工具 btree_mem.go 依赖 GoLLRB，不纳入本包；无新增第三方依赖。
