#!/bin/sh

set -e
set -x

docker build -t geertjohan/aerospike-discovery-manual .

docker push geertjohan/aerospike-discovery-manual:latest
