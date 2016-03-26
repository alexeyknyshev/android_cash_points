#!/bin/bash

SCRIPT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
export GOPATH="$SCRIPT_DIR"
cd "$SCRIPT_DIR/tnt_workdir"
if [[ `pidof tarantool` == "" ]]; then
	echo "run_tests: start tarantool"
	tarantool init.lua &
	TARANTOOL_PID="$!"
fi


cd "$SCRIPT_DIR"

Timeout=15
while [[ `nc -z 0 3302; echo $?` -ne 0 ]]; do
	sleep 1
	Timeout=$((Timeout - 1))
	if [[ Timeout -eq 0 ]]; then
		echo "run_tests: tarantool connection timeout"
		exit 1
	fi
done
if [ $# -ge 1 ]; then
	go test github.com/alexeyknyshev/cpsrv -test.run $1
else
	go test github.com/alexeyknyshev/cpsrv
fi

if [[ $TARANTOOL_PID ]]; then
echo "run_tests: kill tarantool"
[ -e "/proc/$TARANTOOL_PID" ] && kill $TARANTOOL_PID
fi
