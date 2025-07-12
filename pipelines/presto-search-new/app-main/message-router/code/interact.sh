#!/bin/bash

code_dir=`dirname $0`
m=$1
headers=$2

# if message name is new_pointing, then get required info from the headers.

if [[ $m = "update-hosts" ]]; then
    pattern='"volume_low":"([^"]+)"'
    if [[ $headers =~ $pattern ]]; then
        volume_low="${BASH_REMATCH[1]}"
    else
        volume_low=$VOLUME_LOW
    fi

    pattern='"volume_mid":"([^"]+)"'
    if [[ $headers =~ $pattern ]]; then
        volume_mid="${BASH_REMATCH[1]}"
    else
        volume_mid=$VOLUME_MID
    fi

    pattern='"volume_high":"([^"]+)"'
    if [[ $headers =~ $pattern ]]; then
        volume_high="${BASH_REMATCH[1]}"
    else
        volume_high=$VOLUME_HIGH
    fi

    pattern='"node_group":"([^"]+)"'
    if [[ $headers =~ $pattern ]]; then
        node_group="${BASH_REMATCH[1]}"
    else
        node_group="n"
    fi

    /app/bin/update-hosts.py $node_group $volume_low $volume_mid $volume_high

elif [[ $m = "init" ]]; then
    ${code_dir}/init.sh
fi