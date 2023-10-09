# ruijie-telemetry-proto

## protobuf

```
$ cd proto
$ protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative  ruijie-json.proto
```


## example

```
go mod tidy
go build cmd/main.go
```
