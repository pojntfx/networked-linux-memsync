# Set 'e' shell option to exit when a command fails
set -e

# Cleanup
rm -f /tmp/*.benchout

# Set the default number of runs if not provided
RUNS=10

# Loop to run the execution of program
for i in $(seq 1 $RUNS); do
    output=$(/tmp/throughput-disk)
    echo "${output}" >>/tmp/throughput-disk.benchout
done

for i in $(seq 1 $RUNS); do
    output=$(/tmp/throughput-memory)
    echo "${output}" >>/tmp/throughput-memory.benchout
done

for i in $(seq 1 $RUNS); do
    /tmp/benchmark-userfaultfd-go-server &
    server_pid=$!

    sleep 0.1
    output=$(/tmp/throughput-userfaultfd)
    echo "${output}" >>/tmp/benchmark-userfaultfd-go-server.benchout

    kill $server_pid
done

for i in $(seq 1 $RUNS); do
    output=$(/tmp/throughput-r3map-direct)
    echo "${output}" >>/tmp/throughput-r3map-direct.benchout

    sleep 0.3
done

for i in $(seq 1 $RUNS); do
    output=$(/tmp/throughput-r3map-managed)
    echo "${output}" >>/tmp/throughput-r3map-managed.benchout

    sleep 0.3
done

# Remove and recreate output directory
rm -rf bench/throughput-rtt0
mkdir -p bench/throughput-rtt0

# Define headers
echo "Throughput for Disk (0ms RTT) (MB/s),Throughput for Memory (0ms RTT) (MB/s),Throughput for userfaultfd (0ms RTT) (MB/s),Throughput for Direct Mounts (0ms RTT) (MB/s),Throughput for Managed Mounts (0ms RTT) (MB/s)" >bench/throughput-rtt0/results.csv

# Combine outputs
paste -d',' /tmp/throughput-disk.benchout /tmp/throughput-memory.benchout /tmp/benchmark-userfaultfd-go-server.benchout /tmp/throughput-r3map-direct.benchout /tmp/throughput-r3map-managed.benchout >>bench/throughput-rtt0/results.csv

# Cleanup
rm -f /tmp/*.benchout
