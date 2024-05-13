#!/bin/bash
# input message: 1257010784/p00001/group01
m=$1
echo $m

from_ip=$2
echo $from_ip

scalebox task add --sink-job dedisp-search --to-ip $from_ip ${m}