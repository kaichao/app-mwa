#!/usr/local/bin/python3
import os
import sys
from query_db import get_hosts_likely

if __name__ == "__main__":

    # 从环境变量获取 group_id 与 job_id
    group_id = os.getenv("NODES_GROUP")
    print(group_id)
    # 查询数据库，获取 group_id 对应的所有 host
    hosts = get_hosts_likely(group_id)
    # for hh in hosts:
    #     print(hh[0])
    num_hosts = len(hosts)

    if num_hosts == 0:
        sys.exit(1)
    print(hosts)

    with open("./host_list.txt", "w") as f:
        for host in hosts:
            f.write(host[1])
            f.write("\n")
        
        
    
