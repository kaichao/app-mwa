#!/usr/bin/env bash

# /usr/sbin/sshd -f /public/home/cstu0036/.ssh/sshd/sshd_config -D &
/usr/sbin/sshd -f /public/home/cstu0036/.ssh/sshd/sshd_config -D > /dev/null 2>&1 &
SSHD_PID=$!

echo $SSHD_PID
