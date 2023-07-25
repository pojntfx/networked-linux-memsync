# Set 'e' shell option to exit when a command fails
set -e

# Set the default number of runs if not provided
RUNS=5

# Define an array of RTTs
rtts=('0ms' '5ms' '10ms' '15ms' '20ms' '30ms')

# Define an array of pull workers
workers=(0 1 16 512)

# Loop for each RTT
for rtt in ${rtts[@]}; do

    # Loop for each worker
    for worker in ${workers[@]}; do
        for i in $(seq 1 $RUNS); do

            # Run the command for each rtt and worker combination
            /tmp/benchmark-r3map-go-server-managed --pull-workers $worker --rtt $rtt &
            server_pid=$!

            sleep 0.1
            output=$(/tmp/latency-first-chunk-r3map)

            # Construct the output file name
            output_file="/tmp/benchmark-r3map-go-server-managed-${rtt}-${worker}.benchout"
            echo "${output}" >>$output_file

            kill $server_pid
            sleep 0.3
        done
    done
done

# Remove and recreate output directory
rm -rf bench/latency-first-chunk-workervar
mkdir -p bench/latency-first-chunk-workervar

# Define headers
echo "RTT (ms),Worker,First Chunk Latency for Managed Mounts (ns)" >bench/latency-first-chunk-workervar/results.csv

# Combine outputs and add rtt and worker label
for rtt in ${rtts[@]}; do
    for worker in ${workers[@]}; do
        paste -d',' <(printf "%s\n" $rtt) <(printf "%s\n" $worker) /tmp/benchmark-r3map-go-server-managed-$rtt-$worker.benchout >>bench/latency-first-chunk-workervar/results.csv
    done
done

# Cleanup
rm /tmp/*.benchout
