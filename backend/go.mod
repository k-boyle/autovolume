module github.com/k-boyle/autovolume

go 1.27

tool (
	connectrpc.com/connect/cmd/protoc-gen-connect-go
	google.golang.org/protobuf/cmd/protoc-gen-go
)

require (
	connectrpc.com/connect v1.21.0
	google.golang.org/protobuf v1.36.12
)
