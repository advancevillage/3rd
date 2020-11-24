.PHONY: dbproxy cacheProxy clean all

PROTOC:=protoc

all: dbproxy cacheProxy

dbproxy:
	${PROTOC} -I dbproxy/ --go_out=plugins=grpc,paths=source_relative:./dbproxy/ ./dbproxy/*.proto
cacheProxy:
	${PROTOC} -I cacheProxy/ --go_out=plugins=grpc,paths=source_relative:./cacheProxy/ ./cacheProxy/*.proto
clean:
	rm -f dbproxy/*.go
	rm -f cacheProxy/*.go
