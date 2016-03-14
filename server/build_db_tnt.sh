#!/bin/bash

SCRIPT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
#cd "$SCRIPT_DIR/bin"
./bin/server_sqlite_to_tarantool data/towns.db data/cp.db data/banks.db localhost:3301
#cd "$SCRIPT_DIR"

