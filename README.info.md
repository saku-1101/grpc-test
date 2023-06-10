## Commands needed

- `protoc` command installation
```
$ brew install protobuf
```

- Automatic code generation with protoc command
```
protoc --go_out=../pkg/grpc --go_opt=paths=source_relative \
	--go-grpc_out=../pkg/grpc --go-grpc_opt=paths=source_relative \
	hello.proto
```

- `gRPCurl` command installation
```
$ brew install grpcurl
```

- To boot up gRPC server
```
$ go run main.go
```

- To list the available service on the server
```
$ grpcurl -plaintext localhost:8080 list
```

- To list the available method on the service
```
$ grpcurl -plaintext localhost:8080 list [ServiceName]
```

- To send a request to the method
```
$ grpcurl -plaintext -d '{"key": "value"}' localhost:8080 [package].[ServiceName].[MethodName]
```

## Directory Structure
```
./src
├─ api
│   └─ hello.proto # protoファイル
├─ cmd
│   └─ server
│       └─ main.go # gRPC server
├─ pkg
│   └─ grpc # ここにコードを自動生成
├─ go.mod
└─ go.sum
```



