# Set 'e' shell option to exit when a command fails
set -e

# Cleanup
rm -f /tmp/*.benchout

# Set the default number of runs if not provided
RUNS=5

# Define an array of RTTs
rtts=('0ms' '2ms' '4ms' '6ms' '8ms' '10ms' '12ms' '14ms' '22ms' '28ms' '34ms' '40ms')

# Loop for each RTT
for rtt in ${rtts[@]}; do
    echo $rtt

    # Loop to run the execution of program
    for i in $(seq 1 $RUNS); do
        /tmp/benchmark-userfaultfd-go-server --rtt $rtt &
        server_pid=$!

        sleep 0.1
        if [ "$rtt" = '0ms' ]; then
            output=$(/tmp/throughput-userfaultfd)
        else
            # `userfaultfd` is extremely latency sensitive; there is no measureable change
            # in throughput, but a very measureable reduction in execution time if we transfer less data
            output=$(/tmp/throughput-userfaultfd --size 4096)
        fi
        echo "${output}" >>/tmp/benchmark-userfaultfd-go-server-$rtt.benchout

        kill $server_pid
    done

    for i in $(seq 1 $RUNS); do
        if [ "$rtt" = '0ms' ]; then
            output=$(/tmp/throughput-r3map-direct --rtt $rtt)
        else
            # Same as for `userfaultfd`; the same trend persists no matter if we transfer more or less data
            output=$(/tmp/throughput-r3map-direct --size 4194304 --rtt $rtt)
        fi

        echo "${output}" >>/tmp/throughput-r3map-direct-$rtt.benchout

        sleep 0.2
    done

    for i in $(seq 1 $RUNS); do
        # No need to differentiate here since we pull in the background
        output=$(/tmp/throughput-r3map-managed --rtt $rtt)

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
