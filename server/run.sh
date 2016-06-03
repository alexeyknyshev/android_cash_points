#!/bin/bash
declare -a args
test=false
for i in $*; do
	[ $i == "--test" ] && test=true || args[${#args}]=$i
done
if $test ; then
	docker/generator.awk docker/dockerfile_cpsrv_template --test > docker/dockerfile_cpsrv
else
	docker/generator.awk docker/dockerfile_cpsrv_template > docker/dockerfile_cpsrv
fi
set -x
docker-compose -p cashpoints up  $args

