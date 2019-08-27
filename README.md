# protoc-gen-vmicro
v-micro 的 protoc  golang 插件

基于 protoc-gen-go ，实现 v-micro 自定义自动生成代码逻辑


## 编译

- Windows
  ```shell
  build.bat
  ```
- Linux
  ```shell
  ./build.sh
  ```
## 使用方法

```shell
protoc -I. --vmicro_out=. hello.proto
```

需要把 protoc 、 proto-gen-vmicro 拷贝至 hello.proto 所在目录或者系统目录

## broadcast.pb.go 相关

通过 g.bat 、 g.sh 生成至 micro/broadcast.pb.go

依赖以下程序：
- protoc
- protoc-gen-go
