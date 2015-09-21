#!/bin/bash

DATE=`date -u '+%Y/%m/%dT%H:%M:%S'`
go build -ldflags "-X main.BuildDate=$DATE"
