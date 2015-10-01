#!/bin/bash

cd "$( dirname "${BASH_SOURCE[0]}" )"
DATE=`date -u '+%Y/%m/%dT%H:%M:%S'`
go build -ldflags "-X main.BuildDate $DATE"
