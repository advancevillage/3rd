.PHONY: dbproxy cacheProxy clean all

PROTOC:=protoc

all: dbproxy cacheProxy

dbproxy:
	${PROTOC} -I dbproxy/ --go_out=. --go_opt=paths=source_relative  --go-grpc_out=. --go-grpc_opt=paths=source_relative ./dbproxy/*.proto

cacheProxy:
	${PROTOC} -I cacheProxy/ --go_out=. --go_opt=paths=source_relative  --go-grpc_out=. --go-grpc_opt=paths=source_relative ./cacheProxy/*.proto

clean:
	rm -f ./*.pb.go
