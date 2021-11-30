#!/usr/bin/env bash
docker-compose -p lorawan-stack-dev exec -T postgres /bin/bash -c "dropdb --if-exists --force $1 2> /dev/null; createdb $1; psql $1 --quiet" < $2
