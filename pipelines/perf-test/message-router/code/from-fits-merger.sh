#!/bin/bash

filename="/data/mwa/24ch/$1"

# remove 24ch-file
[ "$KEEP_24CH_FILE" = "no" ] && echo "$filename be removed" && rm -f $filename

exit 0
