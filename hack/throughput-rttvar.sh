# Set 'e' shell option to exit when a command fails
set -e

# Cleanup
rm -f /tmp/*.benchout

# Set the default number of runs if not provided
RUNS=2

# Define an array of RTTs
rtts=('0ms' '2ms' '4ms' '6ms' '8ms' '10ms' '12ms' '14ms' '16ms' '18ms' '20ms' '22ms' '24ms' '26ms' '28ms' '30ms' '32ms' '34ms' '36ms' '38ms' '40ms')

# Loop for each RTT
for rtt in ${rtts[@]}; do
    echo $rtt

    # Loop to run the execution of program
    for i in $(seq 1 $RUNS); do
        /tmp/benchmark-userfaultfd-go-server --rtt $rtt &
        server_pid=$!

        sleep 0.1
        output=$(/tmp/throughput-userfaultfd)
        echo "${output}" >>/tmp/benchmark-userfaultfd-go-server-$rtt.benchout

        kill $server_pid
    done

    for i in $(seq 1 $RUNS); do
        output=$(/tmp/throughput-r3map-direct)
        echo "${output}" >>/tmp/throughput-r3map-direct-$rtt.benchout

        sleep 0.2
    done

    for i in $(seq 1 $RUNS); do
        output=$(/tmp/throughput-r3map-managed)
        echo "${output}" >>/tmp/throughput-r3map-managed-$rtt.benchout

        sleep 0.2
    done
done

# Remove and recreate output directory
rm -rf bench/throughput-rttvar
mkdir -p bench/throughput-rttvar

# Define headers
echo "RTT (ms),Throughput for userfaultfd (MB/s),Throughput for Direct Mounts (MB/s),Throughput for Managed Mounts (MB/s)" >bench/throughput-rttvar/results.csv

# Combine outputs and add rtt label
for rtt in ${rtts[@]}; do
    paste -d',' <(printf "%s\n" $rtt) /tmp/benchmark-userfaultfd-go-server-$rtt.benchout /tmp/throughput-r3map-direct-$rtt.benchout /tmp/throughput-r3map-managed-$rtt.benchout >>bench/throughput-rttvar/results.csv
done

# Cleanup
rm -f /tmp/*.benchout
