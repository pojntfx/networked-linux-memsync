# Set 'e' shell option to exit when a command fails
set -e

# Set the default number of runs if not provided
RUNS=5

# Define an array of RTTs
rtts=('0ms' '2ms' '4ms' '6ms' '8ms' '10ms' '12ms' '14ms' '16ms' '18ms' '20ms' '22ms' '24ms' '26ms' '28ms' '30ms' '32ms' '34ms' '36ms' '38ms' '40ms')

# Loop for each RTT
for rtt in ${rtts[@]}; do
    # Loop to run the execution of program
    for i in $(seq 1 $RUNS); do
        /tmp/benchmark-userfaultfd-go-server --rtt $rtt &
        server_pid=$!

        sleep 0.1
        output=$(/tmp/latency-first-chunk-userfaultfd)
        echo "${output}" >>/tmp/benchmark-userfaultfd-go-server-$rtt.benchout

        kill $server_pid
    done

    for i in $(seq 1 $RUNS); do
        /tmp/benchmark-r3map-go-server-direct --rtt $rtt &
        server_pid=$!

        sleep 0.2
        output=$(/tmp/latency-first-chunk-r3map)
        echo "${output}" >>/tmp/benchmark-r3map-go-server-direct-$rtt.benchout

        kill $server_pid
        sleep 0.2
    done

    for i in $(seq 1 $RUNS); do
        /tmp/benchmark-r3map-go-server-managed --rtt $rtt &
        server_pid=$!

        sleep 0.2
        output=$(/tmp/latency-first-chunk-r3map)
        echo "${output}" >>/tmp/benchmark-r3map-go-server-managed-$rtt.benchout

        kill $server_pid
        sleep 0.2
    done
done

# Remove and recreate output directory
rm -rf bench/latency-first-chunk-rttvar
mkdir -p bench/latency-first-chunk-rttvar

# Define headers
echo "RTT (ms),First Chunk Latency for userfaultfd (ns),First Chunk Latency for Direct Mounts (ns),First Chunk Latency for Managed Mounts (ns)" >bench/latency-first-chunk-rttvar/results.csv

# Combine outputs and add rtt label
for rtt in ${rtts[@]}; do
    paste -d',' <(printf "%s\n" $rtt) /tmp/benchmark-userfaultfd-go-server-$rtt.benchout /tmp/benchmark-r3map-go-server-direct-$rtt.benchout /tmp/benchmark-r3map-go-server-managed-$rtt.benchout >>bench/latency-first-chunk-rttvar/results.csv
done

# Cleanup
rm /tmp/*.benchout
