# Set 'e' shell option to exit when a command fails
set -e

# Cleanup
rm -f /tmp/*.benchout

# Set the default number of runs if not provided
RUNS=3

# Define an array of RTTs
rtts=('0ms' '5ms' '10ms' '15ms' '20ms' '30ms')

# Define an array of pull workers
workers=(0 1 16 512 2048 4092 8192 16384)

# Loop for each RTT
for rtt in ${rtts[@]}; do

    # Loop for each worker
    for worker in ${workers[@]}; do
        for i in $(seq 1 $RUNS); do
            echo $rtt $worker $i

            if ([ "$worker" = '0' ] || [ "$worker" = '1' ] || [ "$worker" = '16' ]) && [ "$rtt" != '0ms' ]; then
                # For a very low worker count, the same behavior as for the rttvar benchmark applies, so reduce size to keep execution time reasonable
                output=$(/tmp/throughput-r3map-managed --pull-workers $worker --rtt $rtt --size 4194304)
            else
                output=$(/tmp/throughput-r3map-managed --pull-workers $worker --rtt $rtt)
            fi

            # Construct the output file name
            output_file="/tmp/throughput-r3map-managed-${rtt}-${worker}.benchout"
            echo "${output}" >>$output_file

            sleep 0.3
        done
    done
done

# Remove and recreate output directory
rm -rf bench/throughput-workervar
mkdir -p bench/throughput-workervar

# Define headers
echo "RTT (ms),Worker,Throughput for Managed Mounts (MB/s)" >bench/throughput-workervar/results.csv

# Combine outputs and add rtt and worker label
for rtt in ${rtts[@]}; do
    for worker in ${workers[@]}; do
        paste -d',' <(printf "%s\n" $rtt) <(printf "%s\n" $worker) /tmp/throughput-r3map-managed-$rtt-$worker.benchout >>bench/throughput-workervar/results.csv
    done
done

# Cleanup
rm -f /tmp/*.benchout
