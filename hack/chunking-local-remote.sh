# Set 'e' shell option to exit when a command fails
set -e

# Cleanup
rm -f /tmp/*.benchout

# Set the default number of runs if not provided
RUNS=2

# Define an array of RTTs
rtts=('0ms' '1ms' '2ms' '4ms' '6ms' '10ms' '14ms' '20ms')

# Loop for each RTT
for rtt in ${rtts[@]}; do

    # Loop to run the execution of program
    for i in $(seq 1 $RUNS); do
        echo $rtt direct serverside $i

        /tmp/transport-server --grpc --chunking &
        server_pid=$!

        sleep 0.2

        output=$(/tmp/throughput-r3map-grpc --rtt $rtt)
        echo "${output}" >>/tmp/chunking-local-remote-direct-serverside-$rtt.benchout

        sleep 0.2

        kill $server_pid

        sleep 0.1
    done

    for i in $(seq 1 $RUNS); do
        echo $rtt managed serverside $i

        /tmp/transport-server --grpc --chunking &
        server_pid=$!

        sleep 0.2

        output=$(/tmp/throughput-r3map-grpc --rtt $rtt --managed)
        echo "${output}" >>/tmp/chunking-local-remote-managed-serverside-$rtt.benchout

        sleep 0.2

        kill $server_pid

        sleep 0.1
    done

    for i in $(seq 1 $RUNS); do
        echo $rtt direct clientside $i

        /tmp/transport-server --grpc &
        server_pid=$!

        output=$(/tmp/throughput-r3map-grpc --rtt $rtt --chunking)
        echo "${output}" >>/tmp/chunking-local-remote-direct-clientside-$rtt.benchout

        sleep 0.2

        kill $server_pid

        sleep 0.1
    done

    for i in $(seq 1 $RUNS); do
        echo $rtt managed clientside $i

        /tmp/transport-server --grpc &
        server_pid=$!

        output=$(/tmp/throughput-r3map-grpc --rtt $rtt --chunking --managed)
        echo "${output}" >>/tmp/chunking-local-remote-managed-clientside-$rtt.benchout

        sleep 0.2

        kill $server_pid

        sleep 0.1
    done
done

# Remove and recreate output directory
rm -rf bench/chunking-local-remote
mkdir -p bench/chunking-local-remote

# Define headers
echo "RTT (ms),Throughput for Server-Side Chunking (Direct Mount) (MB/s),Throughput for Server-Side Chunking (Managed Mount) (MB/s),Throughput for Client-Side Chunking (Direct Mount) (MB/s),Throughput for Client-Side Chunking (Managed Mount) (MB/s)" >bench/chunking-local-remote/results.csv

# Combine outputs and add rtt label
for rtt in ${rtts[@]}; do
    paste -d',' <(printf "%s\n" $rtt) /tmp/chunking-local-remote-direct-serverside-$rtt.benchout /tmp/chunking-local-remote-managed-serverside-$rtt.benchout /tmp/chunking-local-remote-direct-clientside-$rtt.benchout /tmp/chunking-local-remote-managed-clientside-$rtt.benchout >>bench/chunking-local-remote/results.csv
done

# Cleanup
rm -f /tmp/*.benchout
