#!/bin/bash

set -ex

CUR_DIR=$PWD
export GOPROXY=https://goproxy.io
export GOBIN=$CUR_DIR/bin

go install ./...

