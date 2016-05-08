#!/bin/bash

echo "docker-compose -p cashpoints up $*"
docker-compose -p cashpoints up $*
