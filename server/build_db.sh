#!/bin/bash

SCRIPT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
cd "$SCRIPT_DIR/bin"
./server_sqlite_to_redis data/towns.db data/cp.db data/banks.db localhost:6379 ./redis_scripts/zclusterdata.lua
cd "$SCRIPT_DIR"
