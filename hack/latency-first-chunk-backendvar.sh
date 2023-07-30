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
        echo $service $i

        output=$(/tmp/latency-first-chunk-r3map-$service)
        echo "${output}" >>/tmp/latency-first-chunk-backendvar-$service.benchout

        sleep 0.5
    done
done

# Remove and recreate output directory
rm -rf bench/latency-first-chunk-backendvar
mkdir -p bench/latency-first-chunk-backendvar

# Define headers
echo "First Chunk Latency for Memory Backend (0ms RTT) (ns),First Chunk Latency for File Backend (0ms RTT) (ns),First Chunk Latency for Directory Backend (0ms RTT) (ns),First Chunk Latency for Redis Backend (0ms RTT) (ns),First Chunk Latency for S3 Backend (0ms RTT) (ns),First Chunk Latency for Cassandra Backend (0ms RTT) (ns)" >bench/latency-first-chunk-backendvar/results.csv

# Combine outputs
paste -d',' /tmp/latency-first-chunk-backendvar-memory.benchout /tmp/latency-first-chunk-backendvar-file.benchout /tmp/latency-first-chunk-backendvar-directory.benchout /tmp/latency-first-chunk-backendvar-redis.benchout /tmp/latency-first-chunk-backendvar-s3.benchout /tmp/latency-first-chunk-backendvar-cassandra.benchout >>bench/latency-first-chunk-backendvar/results.csv

# Cleanup
rm -f /tmp/*.benchout
