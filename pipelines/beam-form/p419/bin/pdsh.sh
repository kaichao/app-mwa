#!/usr/bin/env bash

export PDSH_RCMD_TYPE=ssh
export PDSH_SSH_ARGS_APPEND="-p 50022 -q -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o PreferredAuthentications=publickey"

pdsh -l cstu0030 -w ^/tmp/ip_list.txt $*
