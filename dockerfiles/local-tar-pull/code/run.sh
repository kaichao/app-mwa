#!/bin/bash

# <user>@<ip-addr>/raid0/tmp/mwa/tar1257010784~1257010784/1257010786_1257010815_ch109.dat.tar.zst~/dev/shm/scalebox/mydata/mwa/tar~01
m0=$1

# remove last 3 characters
m="${m0%~*}"

/app/share/bin/run.sh $m

exit $?
