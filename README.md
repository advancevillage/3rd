# 3rd 依赖的第3方库
 * github.com/julienschmidt/httprouter
   * [Document](https://github.com/julienschmidt/httprouter)
 * github.com/gobwas/ws
   * [Document](https://godoc.org/github.com/gobwas/ws)   golang web socket libary
 * github.com/gocql/gocql
   * [Document](https://github.com/gocql/gocql)


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