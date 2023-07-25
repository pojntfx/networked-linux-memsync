#!/bin/bash
set -e

docker rm -f redis && docker run --name redis -p '6379:6379' -d redis
docker rm -f minio && docker run --name minio -p '9000:9000' -d minio/minio server /data
docker rm -f scylladb && docker run --name scylladb -p '9042:9042' -d scylladb/scylla --smp 1
