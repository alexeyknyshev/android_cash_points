#!/bin/bash

# function runTest {
#     testFile=$1
#     currentTestIndex=$2
#     testCount=$3
#     verbose=$4
# 
#     preScriptPath="$testFile.pre.sh"
#     postScriptPath="$testFile.post.sh"
#     host="localhost:$SERVER_PORT"
#     currentTestPrefix="[$currentTestIndex/$testCount] "
# 
#     if [ -e "$preScriptPath" ]
#     then
#         [ $verbose -eq 1 ] && echo "running $preScriptPath"
#         bash "$preScriptPath" "$host"
#     fi
# 
#     [ $verbose -eq 1 ] && echo "${currentTestPrefix}running $testFile"
#     printf "%s" "$currentTestPrefix"
#     pyresttest "$host" "$testFile"
# 
#     if [ -e "$postScriptPath" ]
#     then
#         [ $verbose -eq 1 ] && echo "running $postScriptPath"
#         bash "$postScriptPath" "$host"
#     fi
# }
# 
# verbose=0
# run_direct=""
# if [ $# -ge 1 ]
# then
#     if [ $1 = "-v" ]
#     then
#         verbose=1
#     else
#         run_direct="$1"
#         if [ $# -ge 2 ]
#         then
#             if [ $2 = "-v" ]
#             then
#                 verbose=1
#             fi
#         fi
#     fi
# fi
# 
# SCRIPT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
# cd "$SCRIPT_DIR"
# 
# BIN_DIR="$SCRIPT_DIR/bin"
# CONFIG_FILE='config.json'
# LOG_FILE='server.log'
# SERVER_EXECUTABLE='./server'
# TESTS_DIR='test/unit'
# 
# cd $BIN_DIR
# 
# if [ ! -e "$CONFIG_FILE" ]
# then
#     echo "No such file: config.json"
#     exit 1
# fi
# 
# SERVER_PORT=$(jq '.Port' "$CONFIG_FILE")
# if [ $? -ne 0 ]
# then
#     echo 'Failed to parse server port from "config.json" file'
#     exit 1
# fi
# 
# SERVER_PID=''
# 
# if [ $verbose -eq 0 ]
# then
#     $SERVER_EXECUTABLE "$CONFIG_FILE" >$LOG_FILE 2>$LOG_FILE &
#     SERVER_PID="$!"
# else
#     $SERVER_EXECUTABLE "$CONFIG_FILE" &
#     SERVER_PID="$!"
# fi
# 
# cd $SCRIPT_DIR
# 
# TEST_COUNT=$(find "$TESTS_DIR" -name '*.yaml' | wc -l)
# 
# currentTestIndex=1
# 
# if [[ -z $run_direct ]]
# then
#     find "$TESTS_DIR" -name '*.yaml' | while read testFile
#     do
#         runTest $testFile $currentTestIndex $TEST_COUNT $verbose
#         ((currentTestIndex++))
#     done
# else
#     runTest $run_direct 1 1 $verbose
# fi
# 
# [ -e "/proc/$SERVER_PID" ] && kill $SERVER_PID

SCRIPT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
export GOPATH="$SCRIPT_DIR"

cd "$SCRIPT_DIR/tnt_workdir"
tarantool init.lua &
cd "$SCRIPT_DIR"

sleep 5

go test github.com/alexeyknyshev/cpsrv
