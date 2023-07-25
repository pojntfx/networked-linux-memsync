#!/bin/bash
set -e

# Define array of benchmarks
benchmarks=(
    "latency-first-chunk-disk"
    "latency-first-chunk-memory"
    "latency-first-chunk-r3map"
    "latency-first-chunk-userfaultfd"
    "benchmark-userfaultfd-go-server"
    "benchmark-r3map-go-server-direct"
    "benchmark-r3map-go-server-managed"
    "latency-preemptive-pull"
    "latency-polling-udev"
    "latency-first-chunk-r3map-memory"
    "latency-first-chunk-r3map-file"
    "latency-first-chunk-r3map-directory"
    "latency-first-chunk-r3map-redis"
    "latency-first-chunk-r3map-s3"
    "latency-first-chunk-r3map-cassandra"
    "throughput-memory"
    "throughput-disk"
    "throughput-userfaultfd"
    "throughput-r3map-direct"
    "throughput-r3map-managed"
)

# Loop over benchmarks
for benchmark in "${benchmarks[@]}"; do
    # Build benchmark in the background
    go build -o /tmp/$benchmark ./cmd/$benchmark &
done

# Wait for all background jobs to finish
wait