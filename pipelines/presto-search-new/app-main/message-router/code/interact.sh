#!/bin/bash

code_dir=`dirname $0`
m=$1
headers=$2

# if message name is new_pointing, then get required info from the headers.

if [[ $m = "new_pointing" ]]; then

    pattern='"pointing":"([^"]+)"'
    if [[ $headers =~ $pattern ]]; then
        pointing="${BASH_REMATCH[1]}"
        echo "pointing: $pointing"
    else
        # no from_job in json 
        echo "pointing is not set!" >&2 && exit 20
    fi

    # get the host name from headers.
    pattern='"host":"([^"]+)"'
    if [[ $headers =~ $pattern ]]; then
        host="${BASH_REMATCH[1]}"
        echo "host: $host"
    else
        # no from_job in json 
        echo "host is not set!" >&2 && exit 21
    fi

    sema="pointing-ready:$pointing"
    echo "$sema"
    scalebox semaphore create $sema $NUM_FILES

    # update the semaphore with the host name.
    sema1="host-spare:$host"
    echo "$sema1"
    n=$(scalebox semaphore decrement $sema1)
    code=$?
    if [[ $code -ne 0 ]]; then
        echo "[ERROR] In semaphore decrement:$sema1, ret-code:$code" >&2 && exit 22        
    fi
elif [[ $m = "init" ]]; then
    ${code_dir}/init.sh
fi