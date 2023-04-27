#!/bin/bash

# Loop through each file in the charts directory
for filename in charts/*; do
  # Run the `conftest test` command on the file
  conftest test "$filename"
done