#!/bin/bash

# Extracting relevant information from the input message
# input-message: 1257010784/p00001/t1257010786_1257010845/ch117.fits
# sema: fits-24ch-ready:1257010784/p00001/t1257010786_1257010845
s=$(echo "$1" | cut -d '/' -f 1-3)

sema="fits-24ch-ready:$s"
echo "message:$1,sema:$sema"

# Running the `scalebox` command to get a numeric string
n=$(scalebox semaphore countdown "$sema")
code=$?

# Checking the exit status of the `scalebox` command
# If there is an error, print an error message and exit with the same code
[ $code -ne 0 ] && echo "[ERROR] scalebox semaphore countdown! " >&2 && exit $code 

# Checking if the semaphore is 0
if [ "$n" -eq 0 ]; then
    echo "fits-merger,$s" > $WORK_DIR/messages.txt
fi

echo "finished from-down-sampler." > $WORK_DIR/custom-out.txt

exit 0
