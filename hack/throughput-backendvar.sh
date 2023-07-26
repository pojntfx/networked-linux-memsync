# Set 'e' shell option to exit when a command fails
set -e

# Cleanup
rm -f /tmp/*.benchout

# Set the default number of runs if not provided
RUNS=5

# Loop to run the execution of program
declare -a services=("memory" "file" "directory" "redis" "s3" "cassandra")

for service in "${services[@]}"; do
    for i in $(seq 1 $RUNS); do
        echo direct $service $i

        output=$(/tmp/throughput-r3map-$service --size 4194304) # We can use a lower value here since the reads are linear anyways
        echo "${output}" >>/tmp/throughput-backendvar-direct-$service.benchout

        sleep 0.3
    done

    for i in $(seq 1 $RUNS); do
        echo managed $service $i

        output=$(/tmp/throughput-r3map-$service --size 41943040 --managed)
        echo "${output}" >>/tmp/throughput-backendvar-managed-$service.benchout

        sleep 0.3
    done
done

# Remove and recreate output directory
rm -rf bench/throughput-backendvar
mkdir -p bench/throughput-backendvar

# Define headers
echo "Throughput for Memory Backend (0ms RTT) (Direct Mount) (MB/s),Throughput for File Backend (0ms RTT) (Direct Mount) (MB/s),Throughput for Directory Backend (0ms RTT) (Direct Mount) (MB/s),Throughput for Redis Backend (0ms RTT) (Direct Mount) (MB/s),Throughput for S3 Backend (0ms RTT) (Direct Mount) (MB/s),Throughput for Cassandra Backend (0ms RTT) (Direct Mount) (MB/s),Throughput for Memory Backend (0ms RTT) (Managed Mount) (MB/s),Throughput for File Backend (0ms RTT) (Managed Mount) (MB/s),Throughput for Directory Backend (0ms RTT) (Managed Mount) (MB/s),Throughput for Redis Backend (0ms RTT) (Managed Mount) (MB/s),Throughput for S3 Backend (0ms RTT) (Managed Mount) (MB/s),Throughput for Cassandra Backend (0ms RTT) (Managed Mount) (MB/s)" >bench/throughput-backendvar/results.csv

# Combine outputs
paste -d',' /tmp/throughput-backendvar-direct-memory.benchout /tmp/throughput-backendvar-direct-file.benchout /tmp/throughput-backendvar-direct-directory.benchout /tmp/throughput-backendvar-direct-redis.benchout /tmp/throughput-backendvar-direct-s3.benchout /tmp/throughput-backendvar-direct-cassandra.benchout /tmp/throughput-backendvar-managed-memory.benchout /tmp/throughput-backendvar-managed-file.benchout /tmp/throughput-backendvar-managed-directory.benchout /tmp/throughput-backendvar-managed-redis.benchout /tmp/throughput-backendvar-managed-s3.benchout /tmp/throughput-backendvar-managed-cassandra.benchout >>bench/throughput-backendvar/results.csv

# Cleanup
rm -f /tmp/*.benchout
