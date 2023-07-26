# Set 'e' shell option to exit when a command fails
set -e

# Cleanup
rm -f /tmp/*.benchout

# Set the default number of runs if not provided
RUNS=2

# Define an array of services RTTs
declare -a services=("memory" "file" "directory" "redis" "s3" "cassandra")
rtts=('0ms' '2ms' '4ms' '6ms' '8ms' '10ms' '12ms' '14ms' '22ms' '28ms' '34ms' '40ms')

# Loop for each RTT
for rtt in ${rtts[@]}; do
    # Loop for each service
    for service in "${services[@]}"; do
        # Run the direct and managed tests for each service
        for i in $(seq 1 $RUNS); do
            echo direct $rtt $service $i
            output=$(/tmp/throughput-r3map-$service --rtt $rtt --size 4194304)
            echo "${output}" >>/tmp/throughput-allvar-direct-$service-$rtt.benchout
            sleep 0.3
        done

        for i in $(seq 1 $RUNS); do
            echo managed $rtt $service $i
            output=$(/tmp/throughput-r3map-$service --rtt $rtt --size 41943040 --managed)
            echo "${output}" >>/tmp/throughput-allvar-managed-$service-$rtt.benchout
            sleep 0.3
        done
    done
done

# Remove and recreate output directory
rm -rf bench/throughput-allvar
mkdir -p bench/throughput-allvar

# Define headers
echo "RTT,Service,Throughput for Direct Mounts (MB/s),Throughput for Managed Mounts (MB/s)" >bench/throughput-allvar/results.csv

# Combine outputs and add rtt and service label
for rtt in ${rtts[@]}; do
    for service in "${services[@]}"; do
        paste -d',' <(printf "%s\n" $rtt) <(printf "%s\n" $service) \
            /tmp/throughput-allvar-direct-$service-$rtt.benchout \
            /tmp/throughput-allvar-managed-$service-$rtt.benchout >>bench/throughput-allvar/results.csv
    done
done

# Cleanup
rm -f /tmp/*.benchout
