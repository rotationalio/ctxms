package ctxms

//go:generate protoc -I=./proto --go_out=. --go_opt=module=github.com/rotationalio/ctxms --go-grpc_out=. --go-grpc_opt=module=github.com/rotationalio/ctxms api.proto
