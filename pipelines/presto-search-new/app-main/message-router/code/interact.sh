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

    pattern='"group_size":"([^"]+)"'
    if [[ $headers =~ $pattern ]]; then
        group_size="${BASH_REMATCH[1]}"
    else
        group_size=24
    fi

    /app/bin/update-hosts.py c $group_size $volume_low $volume_mid $volume_high

elif [[ $m = "init" ]]; then
    ${code_dir}/init.sh
fi