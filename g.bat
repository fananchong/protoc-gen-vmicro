set GOPROXY=https://goproxy.io
go get github.com/golang/protobuf
go list -m -f "{{.Dir}}" github.com/golang/protobuf > temp
set /P DEP=<temp
echo %DEP%
cd micro
protoc -I. -I%DEP% --go_out=. broadcast.proto
copy /y github.com\fananchong\protoc-gen-vmicro\micro\broadcast.pb.go .
rd /s /q github.com
cd ..