#!/bin/bash

# Run `go run . chart` and capture output
output=$(go run . chart | grep -v template-kotlin-spring-service | grep -v pipeline-test)

# Loop through each line of output
while read -r line; do
  # Extract the value name from the line
  valuename=$(echo "$line" | awk '{print $1}')

  # Run the `go run . chart render` command
  go run . chart render -o "charts/$valuename.yaml" "$valuename"
done <<< "$output"
