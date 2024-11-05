#!/usr/local/bin/python3
import os
import sys
from query_db import get_same_app_job_by_name, get_job_slot, create_job_slots, get_same_app_jobs, get_hosts

job_slots = {"rfi-find": 1, "dedisp-search": 4, "fold": 1, "result-push": 1}

if __name__ == "__main__":
    # 使用一个字典记录每个job在每个host上应该启动的slot数量

    # 从环境变量获取 group_id 与 job_id
    group_id = os.getenv("NODES_GROUP")
    job_id = os.getenv("JOB_ID")
    # 从命令行参数获取指向参数 pointing
    pointing = int(sys.argv[1])
    # 查询数据库，获取 group_id 对应的所有 host
    hosts = get_hosts(group_id)
    # for hh in hosts:
    #     print(hh[0])
    num_hosts = len(hosts)
    rfi_job = get_same_app_job_by_name(job_id, "rfi-find")
    while(num_hosts > 0):
        # 根据pointing从已获取的host中选出一个
        
        host_index = pointing % num_hosts
        host = hosts[host_index][0]
        print(host)
        rfi_slots = get_job_slot(rfi_job[0][0], host)
        for slot in rfi_slots:
            print(slot)
        # 检查host上rfi-find的slot是否为启动状态
        if len(rfi_slots) > 0 and rfi_slots[0][1] != "OFF":
            break
        # 如果节点上没有启动的slot，为同一app中，job_slots中的所有job创建新的slot
        elif len(rfi_slots) == 0:
            # 为rfi-find创建slot
            create_job_slots(rfi_job[0][0], host, job_slots["rfi-find"])
            # 获取同一app中的其他job
            same_app_jobs = get_same_app_jobs(job_id)
            # 从列表中找到名为dedisp-search，fold，result-push的job，并为其创建slot
            for job in same_app_jobs:
                if job[1] == "dedisp-search":
                    create_job_slots(job[0], host, job_slots["dedisp-search"])
                elif job[1] == "fold":
                    create_job_slots(job[0], host, job_slots["fold"])
                elif job[1] == "result-push":
                    create_job_slots(job[0], host, job_slots["result-push"])
            break

        # 从除去当前指向的host中选择一个
        hosts = hosts[:host_index] + hosts[host_index+1:]
        num_hosts -= 1

    if num_hosts == 0:
        sys.exit(1)
    print(host)
    with open("./host.txt", "w") as f:
        f.write(host)
        
        
    
