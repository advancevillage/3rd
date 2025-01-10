
### 目录说明



### gRPC
1. 安装gRPC核心库
https://pkg.go.dev/google.golang.org/grpc

```
require (
    google.golang.org/grpc v1.69.2
)

```
2. Go plugins for the protocol compiler
```
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

```
3. Protocol buffer compiler, protoc, version 3.
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

