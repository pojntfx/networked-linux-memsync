# Set 'e' shell option to exit when a command fails
set -e

# Cleanup
rm -f /tmp/*.benchout

# Set the default number of runs if not provided
RUNS=2

# Define an array of RTTs
rtts=('0ms' '2ms' '4ms' '6ms' '8ms' '10ms' '14ms' '18ms' '22ms' '28ms' '34ms' '40ms')

# Define an array of protocols
protocols=('dudirekta' 'grpc' 'frpc')

# Loop for each RTT
for rtt in ${rtts[@]}; do
    # Loop for each protocol
    for protocol in ${protocols[@]}; do
        /tmp/transport-server --$protocol &
        server_pid=$!

        sleep 0.2

        # Loop to run the execution of program
        for i in $(seq 1 $RUNS); do
            echo $rtt $protocol direct $i

            output=$(/tmp/throughput-r3map-$protocol --rtt $rtt)
            echo "${output}" >>/tmp/transport-r3map-$protocol-direct-$rtt.benchout

            sleep 0.2
        done

        for i in $(seq 1 $RUNS); do
            echo $rtt $protocol managed $i

            output=$(/tmp/throughput-r3map-$protocol --rtt $rtt --managed)
            echo "${output}" >>/tmp/transport-r3map-$protocol-managed-$rtt.benchout

            sleep 0.2
        done

        kill $server_pid

        sleep 0.1
    done
done

# Remove and recreate output directory
rm -rf bench/rpc-rttvar
mkdir -p bench/rpc-rttvar

# Define headers
echo "RTT (ms),Throughput for dudirekta (Direct Mount) (MB/s),Throughput for dudirekta (Managed Mount) (MB/s),Throughput for grpc (Direct Mount) (MB/s),Throughput for grpc (Managed Mount) (MB/s),Throughput for frpc (Direct Mount) (MB/s),Throughput for frpc (Managed Mount) (MB/s)" >bench/rpc-rttvar/results.csv

# Combine outputs and add rtt label
for rtt in ${rtts[@]}; do
    paste -d',' <(printf "%s\n" $rtt) /tmp/transport-r3map-dudirekta-direct-$rtt.benchout /tmp/transport-r3map-dudirekta-managed-$rtt.benchout /tmp/transport-r3map-grpc-direct-$rtt.benchout /tmp/transport-r3map-grpc-managed-$rtt.benchout /tmp/transport-r3map-frpc-direct-$rtt.benchout /tmp/transport-r3map-frpc-managed-$rtt.benchout >>bench/rpc-rttvar/results.csv
done

# Cleanup
rm -f /tmp/*.benchout
