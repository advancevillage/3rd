.PHONY: dht dbproxy cacheProxy clean all

PROTOC	:=protoc

module 	:= dbproxy cacheProxy dht

all: $(patsubst %,%,$(module))

$(patsubst %,%,$(module)):
	${PROTOC} -I $@/ --go_out=. --go_opt=paths=source_relative  --go-grpc_out=. --go-grpc_opt=paths=source_relative ./$@/*.proto

clean:
	rm -f ./*.pb.go
