# Set 'e' shell option to exit when a command fails
set -e

# Cleanup
rm -f /tmp/*.benchout

# Set the default number of runs if not provided
RUNS=5

# Define an array of RTTs
rtts=('0ms' '1ms' '2ms' '3ms' '4ms' '5ms' '6ms')

# Loop for each RTT
for rtt in ${rtts[@]}; do
    for i in $(seq 1 $RUNS); do
        echo $rtt direct $i

        output=$(/tmp/throughput-r3map-direct-write --rtt $rtt)
        echo "${output}" >>/tmp/throughput-r3map-direct-$rtt.benchout

        sleep 0.2
    done

    for i in $(seq 1 $RUNS); do
        echo $rtt managed $i

        output=$(/tmp/throughput-r3map-managed-write --rtt $rtt)
        echo "${output}" >>/tmp/throughput-r3map-managed-$rtt.benchout

        sleep 0.2
    done
done

# Remove and recreate output directory
rm -rf bench/throughput-write-rttvar
mkdir -p bench/throughput-write-rttvar

# Define headers
echo "RTT (ms),Write Throughput for Direct Mounts (MB/s),Write Throughput for Managed Mounts (MB/s)" >bench/throughput-write-rttvar/results.csv

# Combine outputs and add rtt label
for rtt in ${rtts[@]}; do
    paste -d',' <(printf "%s\n" $rtt) /tmp/throughput-r3map-direct-$rtt.benchout /tmp/throughput-r3map-managed-$rtt.benchout >>bench/throughput-write-rttvar/results.csv
done

# Cleanup
rm -f /tmp/*.benchout
