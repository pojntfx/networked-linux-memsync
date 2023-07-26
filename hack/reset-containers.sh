#!/bin/bash
set -e

docker rm -f redis && docker run --ulimit nofile=262144:262144 --name redis -p '6379:6379' -d redis
docker rm -f minio && docker run --ulimit nofile=262144:262144 --name minio -p '9000:9000' -d minio/minio server /data
docker rm -f scylladb && docker run --ulimit nofile=262144:262144 --name scylladb -p '9042:9042' -d scylladb/scylla --smp 1
