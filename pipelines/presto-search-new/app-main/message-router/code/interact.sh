#!/bin/bash

code_dir=`dirname $0`
m=$1
headers=$2

# if message name is new_pointing, then get required info from the headers.

if [[ $m = "update-hosts" ]]; then
    /app/bin/update-hosts.py c0 24

elif [[ $m = "init" ]]; then
    ${code_dir}/init.sh
fi