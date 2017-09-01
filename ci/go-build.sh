#!/bin/sh

set -e

export GOPATH=$PWD/go
export PATH=$GOPATH/bin:/usr/local/bin:$PATH
export VENDOR_DIR=$PWD/vendor
export RELEASE_DIR=$PWD/release

cd ${GOPATH}/src/${GO_REPO}

cp -R $VENDOR_DIR vendor

GOOS=$BUILD_GOOS GOARCH=$BUILD_GOARCH GO_FLAGS="" make build-all
