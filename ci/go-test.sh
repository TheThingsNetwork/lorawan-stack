#!/bin/sh

set -e

export GOPATH=$PWD/go
export PATH=$GOPATH/bin:/usr/local/bin:$PATH
export VENDOR_DIR=$PWD/vendor

cd ${GOPATH}/src/${GO_REPO}

cp -R $VENDOR_DIR vendor

GOARCH=${TEST_GOARCH} make go.test
