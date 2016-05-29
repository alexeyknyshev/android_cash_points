#!/bin/bash

SCRIPT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
./bin/server_sqlite_to_tarantool data/towns.db data/cp.db data/banks.db localhost:3301

