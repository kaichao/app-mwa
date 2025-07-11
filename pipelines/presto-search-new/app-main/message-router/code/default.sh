#!/bin/bash

# check the input format.
code_dir=`dirname $0`
m0=$1
headers=$2

# message will be in 3 formats:
# 1. 1257010784/p00001_00024
# 2. 1257010784/p00023/t1257010786_1257010965
# 3. Command:XXXXXX

echo "message in default: $m0"
# now check the input format.
if [[ $m0 =~ ^([^/]+)/p([0-9]+)_([0-9]+)$ ]]; then
    # first format
    ${code_dir}/shared_pointings.sh $m0 $headers
elif [[ $m0 =~ ^([^/]+)/p([0-9]+)$ ]]; then
    # second format
    # execute to_fits-merge.sh
    ${code_dir}/single_pointing.sh $m0 $headers
elif [[ $m0 =~ ^Command:(.+)$ ]]; then
    # third format
    m=${BASH_REMATCH[1]}
    # execute interact.sh
    ${code_dir}/interact.sh $m $headers
elif [[ $m0 =~ ^recover:([^/]+)/p([0-9]+)$ ]]; then
    dataset=${BASH_REMATCH[1]}
    pointing=${BASH_REMATCH[2]}
    ${code_dir}/recover_pointing.py ${dataset}/${pointing}
else
    echo "[ERROR] In checking input format:$m0" >&2 && exit 1
fi
