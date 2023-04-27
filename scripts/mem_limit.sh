#!/bin/bash

set -e

# Set the directory containing the YAML files
YAML_DIR="charts"

# Initialize the total memory limit and total memory request
total_memory_limit="0"
total_memory_request="0"
total_cpu_limit="0"
total_cpu_request="0"

printf "%-30s %-15s %-15s %-15s %-15s\n" "SERVICE" "CREQUEST" "CLIMIT" "MREQUEST" "MLIMIT"
# Loop through each YAML file in the directory
for file in $YAML_DIR/*.yaml
do
    # Get the filename without extension
    filename=$(basename "$file" .yaml)

    # Check if the file contains a Deployment
    if grep -q "kind: Deployment" "$file"; then
        # Get the memory limit and request values (if any)
        memory_request=$(yq -r '.spec.template.spec.containers[].resources.requests.memory' "$file")
        memory_limit=$(yq -r '.spec.template.spec.containers[].resources.limits.memory' "$file")
        cpu_request=$(yq -r '.spec.template.spec.containers[].resources.requests.cpu' "$file")
        cpu_limit=$(yq -r '.spec.template.spec.containers[].resources.limits.cpu' "$file")

        # Add the memory limit and request to the total (if not null)
        if [ -n "$memory_limit" ]; then
            memory_limit_ki=$(echo "$memory_limit" | sed 's/Mi$/ * 1024/' | sed 's/Gi$/ * 1024 * 1024/' | bc)
            total_memory_limit=$(echo "$total_memory_limit + $memory_limit_ki" | bc)
        fi

        if [ -n "$memory_request" ]; then
            memory_request_ki=$(echo "$memory_request" | sed 's/Mi$/ * 1024/' | sed 's/Gi$/ * 1024 * 1024/' | bc)
            total_memory_request=$(echo "$total_memory_request + $memory_request_ki" | bc)
        fi
        
        if [ -n "$cpu_limit" ]; then
            cpu_limit_ki=$(echo "$cpu_limit" | sed 's/Mi$/ * 1024/' | sed 's/m$/ * 1024/' | sed 's/Gi$/ * 1024 * 1024/' | sed 's/MB$/ * 1000 * 1024/' | bc)
            total_cpu_limit=$(echo "$total_cpu_limit + $cpu_limit_ki" | bc)
        fi

        if [ -n "$cpu_request" ]; then
            cpu_request_ki=$(echo "$cpu_request" | sed 's/Mi$/ * 1024/' | sed 's/m$/ * 1024/' | sed 's/Gi$/ * 1024 * 1024/' | sed 's/MB$/ * 1000 * 1024/' | bc)
            total_cpu_request=$(echo "$total_cpu_request + $cpu_request_ki" | bc)
        fi

        # Print the result in a table
        printf "%-30s %-15s %-15s %-15s %-15s\n" "$filename" "${cpu_request:-No request}" "${cpu_limit:-No limit}" "${memory_request:-No request}" "${memory_limit:-No limit}"
    else
        # Print a message if the file doesn't contain a Deployment
        printf "%-30s %s\n" "$filename" "No Deployment found"
    fi
done

# Convert total memory limit and request from Ki to Mi and print them
total_memory_limit=$(echo "$total_memory_limit / 1000" | bc)
total_memory_request=$(echo "$total_memory_request / 1000" | bc)
printf "Total CPU request: %s\n" "${total_cpu_request:-No request}"
printf "Total CPU limit: %s\n" "${total_cpu_limit:-No limit}"
printf "Total memory request: %sMi\n" "${total_memory_request:-No request}"
printf "Total memory limit: %sMi\n" "${total_memory_limit:-No limit}"

