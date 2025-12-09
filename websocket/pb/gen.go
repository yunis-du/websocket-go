package pb

//go:generate protoc --gogo_out=. -I. -I$GOPATH/pkg/mod/github.com/gogo/protobuf@v1.3.2 ./structmessage.proto
