#!/bin/bash
set -e

podman rm -f redis && podman run --ulimit nofile=262144:262144 --name redis -p '6379:6379' -d redis
podman rm -f minio && podman run --ulimit nofile=262144:262144 --name minio -p '9000:9000' -d minio/minio server /data
podman rm -f scylladb && podman run --ulimit nofile=262144:262144 --name scylladb -p '9042:9042' -d scylladb/scylla --smp 1
