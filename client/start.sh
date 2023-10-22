#!/bin/bash

count=150

for ((i=1; i<=count; i++))
do
    echo "Starting client $i"
    "./main" &
    pids[${i}]=$!
done

for pid in ${pids[*]}
do
    wait $pid
done

echo "All clients finished."