# Set 'e' shell option to exit when a command fails
set -e

# Cleanup
rm -f /tmp/*.benchout

# Set the default number of runs if not provided
RUNS=100

# Loop to run the execution of program
for i in $(seq 1 $RUNS); do
    output=$(/tmp/latency-polling-udev)
    echo "${output}" >>/tmp/direct-mount-initialization-time-polling.benchout

    sleep 0.3
done

for i in $(seq 1 $RUNS); do
    output=$(/tmp/latency-polling-udev --udev)
    echo "${output}" >>/tmp/direct-mount-initialization-time-udev.benchout

    sleep 0.3
done

# Remove and recreate output directory
rm -rf bench/latency-polling-udev
mkdir -p bench/latency-polling-udev

# Define headers
echo "Direct Mount Initialization Time (Polling) (ns),Direct Mount Initialization Time (udev) (ns)" >bench/latency-polling-udev/results.csv

# Combine outputs
paste -d',' /tmp/direct-mount-initialization-time-polling.benchout /tmp/direct-mount-initialization-time-udev.benchout >>bench/latency-polling-udev/results.csv

# Cleanup
rm -f /tmp/*.benchout
