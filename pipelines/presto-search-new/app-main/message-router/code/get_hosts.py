#!/usr/local/bin/python3
import os
import sys
from query_db import get_same_app_job_by_name, get_job_slot, create_job_slots, get_same_app_jobs, get_hosts, get_hosts_likely_ordered

job_slots = {"local-copy-unpack": 1,"rfi-find": 1, "dedisp-search": 4, "fold": 1, "result-push": 1}

if __name__ == "__main__":
    # 使用一个字典记录每个job在每个host上应该启动的slot数量

    # 从环境变量获取 group_id 与 job_id
    group_id = os.getenv("NODES_GROUP")
    job_id = os.getenv("JOB_ID")
    # 从命令行参数获取指向参数 pointing
    pointing = int(sys.argv[1])
    rfi_job = get_same_app_job_by_name(job_id, "rfi-find")
    unpack_job = get_same_app_job_by_name(job_id, "local-copy-unpack")
    dedisp_job = get_same_app_job_by_name(job_id, "dedisp-search")
    print(unpack_job[0][0])
    print(dedisp_job[0][0])
    # 查询数据库，获取 group_id 对应的所有 host
    # 任务中有rfi_job时使用。否则置为-1。
    if (len(rfi_job) > 0):
        rfi_id = rfi_job[0][0]
    else:
        rfi_id = -1
    hosts = get_hosts_likely_ordered(rfi_id, unpack_job[0][0], dedisp_job[0][0], group_id)
    # for hh in hosts:
    #     print(hh[0])
    num_hosts = len(hosts)

    idx = 0
    while(idx < num_hosts):
        # 已根据待处理指向数量排序。选择第一个
        host = hosts[idx][0]
        allocated = hosts[idx][1]
        unpack_slots = get_job_slot(unpack_job[0][0], host)
        print(host, allocated)
        # for slot in rfi_slots:
        #     print(slot)
        # 检查host上local-copy-unpack的slot是否为启动状态
        if len(unpack_slots) > 0 and unpack_slots[0][1] != "OFF":
            break
        # 如果节点上没有启动的slot，为同一app中，job_slots中的所有job创建新的slot
        elif len(unpack_slots) == 0:
            # 为local-copy-unpack创建slot
            create_job_slots(unpack_job[0][0], host, job_slots["rfi-find"])
            # 获取同一app中的其他job
            same_app_jobs = get_same_app_jobs(job_id)
            # 从列表中找到名为rfi-find,dedisp-search，fold，result-push的job，并为其创建slot
            for job in same_app_jobs:
                if job[1] == "rfi-find":
                    create_job_slots(job[0], host, job_slots["rfi-find"])
                elif job[1] == "dedisp-search":
                    create_job_slots(job[0], host, job_slots["dedisp-search"])
                elif job[1] == "fold":
                    create_job_slots(job[0], host, job_slots["fold"])
                elif job[1] == "result-push":
                    create_job_slots(job[0], host, job_slots["result-push"])
            break
        # 如果slot已停止（节点即将释放），选下一个
        idx += 1

    if num_hosts == 0:
        sys.exit(1)
    # print(host)
    with open("./host.txt", "w") as f:
        f.write(host)
        
        
    
