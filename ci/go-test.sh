#!/bin/bash

set -e

# Kill background jobs on exit
trap 'kill $(jobs -p)' EXIT

if [[ ! -d ./deps ]]
then
  mkdir -p ./deps
fi

if [[ ! -f ./deps/cockroach-v1.0.6 ]]
then
  wget -qO- https://binaries.cockroachdb.com/cockroach-v1.0.6.linux-amd64.tgz | tar xvz
  mv ./cockroach-v1.0.6.linux-amd64/cockroach ./deps/cockroach-v1.0.6
  chmod +x ./deps/cockroach-v1.0.6
fi
./deps/cockroach-v1.0.6 start --insecure &

if [[ ! -f ./deps/redis-v4.0.2 ]]
then
  wget -qO- http://download.redis.io/releases/redis-4.0.2.tar.gz | tar xvz
  pushd redis-4.0.2/src > /dev/null
  make redis-server
  popd > /dev/null
  mv ./redis-4.0.2/src/redis-server ./deps/redis-v4.0.2
fi
./deps/redis-v4.0.2 &

export GOPATH=$PWD/go
export PATH=$GOPATH/bin:/usr/local/bin:$PATH
export VENDOR_DIR=$PWD/vendor

cd ${GOPATH}/src/${GO_REPO}

cp -R $VENDOR_DIR vendor

GOARCH=${TEST_GOARCH} make go.test
