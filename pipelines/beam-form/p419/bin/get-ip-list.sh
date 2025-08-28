#!/usr/bin/env bash

docker exec -i database psql -Uscalebox -t -A -P pager=off > /tmp/ip_list.txt << EOF
  SELECT ip_addr FROM t_host WHERE hostname ~ 'd[0-9]+.+p419' AND status='ON' ORDER BY hostname
EOF

scp /tmp/ip_list.txt login2:/tmp/ip_list_30.txt
