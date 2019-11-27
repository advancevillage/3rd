### feature
| 功能    |  单元测试 | 基准测试  |
| :-----:| :----:   | :----:  |
| rcp    |          |         |
| http   |    Y     |         |
| queue  |          |         |
| notice |          |         |
| logs   |          |         |
| ws     |          |         |
| pool   |          |         |
| times  |          |         |
| utils  |          |         |
| files  |          |         |


#### init
````shell
    export GOPROXY=https://goproxy.io
    export GO111MODULE=on
````

#### go mod
````shell
    //清缓存
    $ go clean -modcache

    go.mod：依赖列表和版本约束。
    go.sum：记录module文件hash值，用于安全校验
````

#### question
````md
    q: $GOPATH/go.mod exists but should not
    a: mod 不能与gopath共存
````