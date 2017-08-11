#!/bin/bash

go get -d -u github.com/grpc-ecosystem/grpc-gateway/...

sed \
  -e 's:runtime:jsonpb:' \
  -e 's:golang/protobuf:gogo/protobuf:g' \
  -e 's:JSONPb:GoGoJSONPb:g' \
  -e '/import (/a \
  . "github.com/grpc-ecosystem/grpc-gateway/runtime"' \
  $GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/runtime/marshal_jsonpb.go > marshaling.go

goimports -w marshaling.go

echo -e "// Code modified from github.com/grpc-ecosystem/grpc-gateway/runtime by copy.sh\n\n$(cat marshaling.go)" > marshaling.go
