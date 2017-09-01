#!/bin/sh

set -e

export GOPATH=$PWD/go
export PATH=$GOPATH/bin:/usr/local/bin:$PATH
export VENDOR_DIR=$PWD/vendor

cd ${GOPATH}/src/${GO_REPO}

export TMPDIR=$PWD/.tmp
mkdir -p $TMPDIR

make go.dev-deps

dep ensure -v -vendor-only

cp -R vendor/* $VENDOR_DIR
