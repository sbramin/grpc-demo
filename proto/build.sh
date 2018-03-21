#!/bin/sh

repo="$( cd "$( dirname "$1" )" && pwd )"

# Install required repos if you don't have them.
[ -d $GOPATH/src/google.golang.org/grpc ] || go get google.golang.org/grpc

[ -d $GOPATH/bin/protoc-min-version ] || go get github.com/gogo/protobuf/protoc-min-version 
[ -d $GOPATH/bin/protoc-gen-gofast ] || go get github.com/gogo/protobuf/protoc-gen-gofast

[ -d $GOPATH/bin/protoc-gen-grpc-gateway ] || go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway
[ -d $GOPATH/bin/protoc-gen-swagger ] || go get github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger


# Proto Path
protoPath="$repo/proto"

# Go PB Path
goout=$repo/pkg/pb

# Import for gogoproto
gogo="-I=$GOPATH/src"

# Include for gogoproto timestmap
#timestamp="Mgoogle/protobuf/timestamp.proto=github.com/gogo/protobuf/types"

[ -d $goout ] || mkdir -p $goout


# Generate Swagger for service
$GOPATH/bin/protoc-min-version --version="3.0.0" $gogo -I$protoPath/ -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis\
	--swagger_out=logtostderr=true:/tmp $protoPath/service.proto

# Generate GRPC gateway
$GOPATH/bin/protoc-min-version --version="3.0.0" $gogo -I$protoPath/ -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis\
	--grpc-gateway_out=logtostderr=true:$GOPATH/src $protoPath/service.proto

# Generate GRPC service
protoc-min-version --version="3.0.0" $gogo -I$protoPath/ -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis --gofast_out=plugins=grpc:${GOPATH}/src $protoPath/service.proto


