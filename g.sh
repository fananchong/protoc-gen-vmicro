#!/bin/bash

set -ex

export GOPROXY=https://goproxy.io
go get github.com/golang/protobuf
DEP=`go list -m -f "{{.Dir}}" github.com/golang/protobuf`
echo $DEP
cd micro
protoc -I. -I$DEP --go_out=. broadcast.proto
cp -f github.com/fananchong/protoc-gen-vmicro/micro/broadcast.pb.go .
rm -rf github.com
cd ..

