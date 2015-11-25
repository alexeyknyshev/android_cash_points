#!/bin/bash

SCRIPT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$SCRIPT_DIR"
DATE=`date -u '+%Y/%m/%dT%H:%M:%S'`
export GOPATH="$SCRIPT_DIR"
#go get
go build github.com/alexeyknyshev/server
go build github.com/alexeyknyshev/tools/server_sqlite_to_redis
go install github.com/alexeyknyshev/server
go install github.com/alexeyknyshev/tools/server_sqlite_to_redis
rm "$SCRIPT_DIR"/server
