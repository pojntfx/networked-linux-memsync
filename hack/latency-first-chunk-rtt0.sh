# Set 'e' shell option to exit when a command fails
set -e

# Cleanup
rm -f /tmp/*.benchout

# Set the default number of runs if not provided
RUNS=100

# Loop to run the execution of program
for i in $(seq 1 $RUNS); do
    output=$(/tmp/latency-first-chunk-disk)
    echo "${output}" >>/tmp/latency-first-chunk-disk.benchout
done

for i in $(seq 1 $RUNS); do
    output=$(/tmp/latency-first-chunk-memory)
    echo "${output}" >>/tmp/latency-first-chunk-memory.benchout
done

for i in $(seq 1 $RUNS); do
    /tmp/benchmark-userfaultfd-go-server &
    server_pid=$!

    sleep 0.1
    output=$(/tmp/latency-first-chunk-userfaultfd)
    echo "${output}" >>/tmp/benchmark-userfaultfd-go-server.benchout

    kill $server_pid
done

for i in $(seq 1 $RUNS); do
    /tmp/benchmark-r3map-go-server-direct &
    server_pid=$!

    sleep 0.2
    output=$(/tmp/latency-first-chunk-r3map)
    echo "${output}" >>/tmp/benchmark-r3map-go-server-direct.benchout

    kill $server_pid
    sleep 0.2
done

for i in $(seq 1 $RUNS); do
    /tmp/benchmark-r3map-go-server-managed &
    server_pid=$!

    sleep 0.2
    output=$(/tmp/latency-first-chunk-r3map)
    echo "${output}" >>/tmp/benchmark-r3map-go-server-managed.benchout

    kill $server_pid
    sleep 0.2
done

# Remove and recreate output directory
rm -rf bench/latency-first-chunk-rtt0
mkdir -p bench/latency-first-chunk-rtt0

# Define headers
echo "First Chunk Latency for Disk (0ms RTT) (ns),First Chunk Latency for Memory (0ms RTT) (ns),First Chunk Latency for userfaultfd (0ms RTT) (ns),First Chunk Latency for Direct Mounts (0ms RTT) (ns),First Chunk Latency for Managed Mounts (0ms RTT) (ns)" >bench/latency-first-chunk-rtt0/results.csv

# Combine outputs
paste -d',' /tmp/latency-first-chunk-disk.benchout /tmp/latency-first-chunk-memory.benchout /tmp/benchmark-userfaultfd-go-server.benchout /tmp/benchmark-r3map-go-server-direct.benchout /tmp/benchmark-r3map-go-server-managed.benchout >>bench/latency-first-chunk-rtt0/results.csv

# Cleanup
rm -f /tmp/*.benchout
