# Set 'e' shell option to exit when a command fails
set -e

# Cleanup
rm -f /tmp/*.benchout

# Set the default number of runs if not provided
RUNS=5

# Define an array of RTTs
rtts=('0ms' '1ms' '2ms' '3ms' '5ms' '7ms' '10ms' '15ms' '20ms')

# Define an array of pull workers
workers=(1 64 512 2048 4096)

# Loop for each RTT
for rtt in ${rtts[@]}; do

    # Loop for each worker
    for worker in ${workers[@]}; do
        for i in $(seq 1 $RUNS); do
            # Run the command for each rtt and worker combination
            output=$(/tmp/latency-preemptive-pull --pull-workers $worker --rtt $rtt)

            # Construct the output file name
            output_file="/tmp/benchmark-r3map-go-server-managed-${rtt}-${worker}.benchout"
            echo "${output}" >>$output_file

            sleep 0.3
        done
    done
done

# Remove and recreate output directory
rm -rf bench/latency-preemptive-pull
mkdir -p bench/latency-preemptive-pull

# Define headers
echo "RTT (ms),Worker,Preemptive Pulls for Managed Mounts (Byte)" >bench/latency-preemptive-pull/results.csv

# Combine outputs and add rtt and worker label
for rtt in ${rtts[@]}; do
    for worker in ${workers[@]}; do
        paste -d',' <(printf "%s\n" $rtt) <(printf "%s\n" $worker) /tmp/benchmark-r3map-go-server-managed-$rtt-$worker.benchout >>bench/latency-preemptive-pull/results.csv
    done
done

# Cleanup
rm -f /tmp/*.benchout
