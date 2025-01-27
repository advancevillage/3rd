# 想你所想，无限可能

## 目录
- [第一章：gRPC](#gRPC)
- [第二章：证书](#申请证书)

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

## mariadb
1. 安装
```
docekr pull mariadb:11.6.2
```
2. 启动
```
docker run -d --name maria  -e MARIADB_ROOT_PASSWORD=password -p 3306:3306 mariadb:11.6.2
```

## redis
1. 安装
```
docekr pull redis:latest
```
2. 启动
```
docker run -d --name redis -p 6379:6379 redis:latest
```


