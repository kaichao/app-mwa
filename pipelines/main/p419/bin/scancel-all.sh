#!/usr/bin/env bash

for j in $(squeue |grep -v JOBID|awk '{print $1}'); do echo $j;scancel $j;done
