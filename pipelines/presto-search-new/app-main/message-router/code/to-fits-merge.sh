#!/bin/bash

# parse the host from other header name...

m=$1
headers=$2

pattern='"start_host":"([^"]+)"'
if [[ $headers =~ $pattern ]]; then
    start_host="${BASH_REMATCH[1]}"
    echo "start_host: $start_host"
else
    # no from_job in json 
    start_host=""
fi

scalebox task add --sink-job=local-wait-queue --to-host $start_host $m