#!/bin/bash

set -e

if [[ ! -f ./deps/cockroach-v1.0.6 ]]
then
  wget -qO- https://binaries.cockroachdb.com/cockroach-v1.0.6.linux-amd64.tgz | tar xvz
  mkdir -p ./deps
  mv ./cockroach-v1.0.6.linux-amd64/cockroach ./deps/cockroach-v1.0.6
  chmod +x ./deps/cockroach-v1.0.6
fi
./deps/cockroach-v1.0.6 start --insecure &

export GOPATH=$PWD/go
export PATH=$GOPATH/bin:/usr/local/bin:$PATH
export VENDOR_DIR=$PWD/vendor

cd ${GOPATH}/src/${GO_REPO}

cp -R $VENDOR_DIR vendor

GOARCH=${TEST_GOARCH} make go.test
