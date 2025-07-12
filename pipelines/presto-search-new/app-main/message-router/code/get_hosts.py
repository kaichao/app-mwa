#!/usr/local/bin/python3
import os
import sys
from query_db import get_same_app_job_by_name, get_job_slot, get_same_app_jobs, get_hosts, get_hosts_likely_ordered

if __name__ == "__main__":
    # 使用一个字典记录每个job在每个host上应该启动的slot数量

    # 从环境变量获取 group_id 与 job_id
    group_id = os.getenv("NODES_GROUP")
    job_id = os.getenv("JOB_ID")
    # 从命令行参数获取指向参数 pointing
    pointing = int(sys.argv[1])
    rfi_job = get_same_app_job_by_name(job_id, "rfi-find")
    copy_job = get_same_app_job_by_name(job_id, "local-copy")
    unpack_job = get_same_app_job_by_name(job_id, "local-copy-unpack")
    dedisp_job = get_same_app_job_by_name(job_id, "dedisp-search")
    print(copy_job[0][0])
    print(dedisp_job[0][0])
    # 查询数据库，获取 group_id 对应的所有 host
    # 任务中有rfi_job时使用。否则置为-1。
    if (len(rfi_job) > 0):
        rfi_id = rfi_job[0][0]
    else:
        rfi_id = -1
    hosts = get_hosts_likely_ordered(copy_job[0][0], unpack_job[0][0], dedisp_job[0][0], group_id)
    # for hh in hosts:
    #     print(hh[0])
    num_hosts = len(hosts)

    idx = 0
    while(idx < num_hosts):
        # 已根据待处理指向数量排序。选择第一个
        host = hosts[idx][0]
        hostname = hosts[idx][1]
        allocated = hosts[idx][2]
        copy_slots = get_job_slot(copy_job[0][0], host)
        print(host, hostname, allocated)
        # for slot in rfi_slots:
        #     print(slot)
        # 检查host上local-unpack的slot是否为启动状态
        if len(copy_slots) > 0 and copy_slots[0][1] != "OFF":
            break
        # 不再处理新增的节点，仅检查是否可用。
        # 如果slot已停止（节点即将释放），选下一个
        idx += 1

    if num_hosts == 0:
        sys.exit(1)
    # print(host)
    with open("./host.txt", "w") as f:
        f.write(hostname)
        
        
    
