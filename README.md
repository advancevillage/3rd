# 想你所想，无限可能

## 目录
- [gRPC](#gRPC)
- [证书](#申请证书)
- [数据库](#数据库)
- [发布/订阅](#发布/订阅)
- [分布式锁](#分布式锁)
- [JWT](#JWT)
- [单元测试](#单元测试)

## gRPC
1. 安装grpc核心库
```
require (
    google.golang.org/grpc v1.69.2
)

```
2. go grpc 核心
```
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```
3. pb编译器
```
brew install protobuf (MacOS)
apt install -y protobuf-compiler (Linux)

$ protoc --version
$ libprotoc 3.12.4
```
4. 设置PATH
```
export PATH="$PATH:$(go env GOPATH)/bin"
```
5. 编辑proto
```
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative netx/grpc.proto
```
6. 文档
https://grpc.io/docs/languages/go/quickstart/


## 申请证书
1. 下载Certbot
```
brew install certbot
certbot --version
```
地址：https://certbot.eff.org/instructions?ws=nginx&os=osx

2. 获取证书
```
certbot certonly --manual --preferred-challenges dns -d *.sunhe.org -d sunhe.org --server https://acme-v02.api.letsencrypt.org/directory     

可能问题:
q. SSL: CERTIFICATE_VERIFY_FAILED]
a. pip3 install --upgrade certifi
```
3. 域名添加记录
```
添加: txt _acme-challenge.sunhe.org xxxxxxx
生效: dig txt _acme-challenge.sunhe.org @8.8.8.8
```

## 数据库
1. 安装
```
docker pull mariadb:11.6.2
```
2. 启动
```
docker run -d --name maria  -e MARIADB_ROOT_PASSWORD=password -p 3306:3306 mariadb:11.6.2
```

## 发布/订阅
基于Redis Stream 实现Pub/Sub
1. 安装
```
docekr pull redis:latest
```
2. 启动
```
docker run -d --name redis -p 6379:6379 redis:latest
```

## 分布式锁
1. 操作的原子性
```
get and compare/read and save 等操作都是非原子性的，需要借助Lua脚本实现。

```
2. 谁加锁谁解锁
```
set key value [EX seconds] [PX milliseconds] [NX|XX] 虽然原子操作，实际并不安全，当执行耗时>过期事件。
go1:  加锁-->CPU抖动-->过期释放-->执行中-->释放锁
go2:                   加锁-->执行中-->释放锁(已被释放)
go3:                          加锁-->执行中-->释放锁(已被释放)
gon:                                  加锁-->执行中-->释放锁(已被释放)
形成雪崩效应。所以谁加锁谁解锁。
```
3. 集群延迟保护
```
TODO
```

## JWT
https://github.com/golang-jwt/jwt v5.2.1

## 单元测试
1. 数据库
```
dlv test ./dbx -- -test.run ^TestMariaSqlExecutor_ExecSql$
b dbx/sqlx_test.go:107 
```
2. 网络
```
go test -v -count=1 -test.run Test_http ./netx/
```
Q: 代理（Proxy）或 VPN 导致 TLS 连接失败
A: 关闭代理; 进入Postman > Settings（⚙️ 设置）> Proxy), 关闭 "Use System Proxy"


## 连通性
1. nc -zv 127.0.0.1 6379
