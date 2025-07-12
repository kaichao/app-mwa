#!/usr/local/bin/python3

import os
import sys
from query_db import get_same_app_job_by_name, get_task_by_job, get_host_by_ip

def main():
    # 从环境变量获取job_id
    job_id = os.environ.get('JOB_ID')
    if not job_id:
        sys.stderr.write("[ERROR] JOB_ID not set\n")
        return 1
    
    # 从命令行参数获取from_ip
    if len(sys.argv) < 2:
        sys.stderr.write("[ERROR] need more arguments\n")
        return 1
    
    from_ip = sys.argv[1]
    
    # 1. 获取当前JOB相同app的local-copy-unpack和local-wait-queue的job_id
    local_copy_unpack_jobs = get_same_app_job_by_name(job_id, "local-copy-unpack")
    local_copy_jobs = get_same_app_job_by_name(job_id, "local-copy")
    local_wait_queue_jobs = get_same_app_job_by_name(job_id, "local-wait-queue")
    
    # 检查是否找到了相应的job
    if not local_copy_unpack_jobs or not local_copy_jobs:
        sys.stderr.write("[ERROR] no relative job\n")
        return 1
    
    # 提取job_id
    local_copy_unpack_job_id = local_copy_unpack_jobs[0][0]
    local_copy_job_id = local_copy_jobs[0][0]
    local_wait_queue_job_id = local_wait_queue_jobs[0][0]
    
    # 获取from_ip对应的hostname
    host_rows = get_host_by_ip(from_ip)
    if not host_rows:
        sys.stderr.write(f"[ERROR] cannot find host {from_ip}\n")
        return 1
    
    from_host = host_rows[0][0]
    
    # 2. 获取local-copy-unpack在from_host上的状态为-1的任务数量
    local_copy_unpack_count = 0
    local_copy_unpack_tasks = get_task_by_job(local_copy_unpack_job_id, -1, from_host)
    if local_copy_unpack_tasks:
        local_copy_unpack_count = local_copy_unpack_tasks[0][0]
    
    # 获取local-copy在from_host上状态为-1的任务数量
    local_copy_count = 0
    local_copy_tasks = get_task_by_job(local_copy_job_id, -1, from_host)
    if local_copy_tasks:
        local_copy_count = local_copy_tasks[0][0]

    # 获取local-wait-queue的状态为-1的总任务数量
    local_wait_queue_count = 0
    local_wait_queue_tasks = get_task_by_job(local_wait_queue_job_id, -1)
    if local_wait_queue_tasks:
        local_wait_queue_count = local_wait_queue_tasks[0][0]
    
    # 3. 如果本节点无任务，且local-copy有任务，返回1
    # 4. 否则返回0
    sys.stderr.write(f"[INFO] {local_copy_unpack_count} local pointings waiting, {local_copy_count} pointings ready, {local_wait_queue_count} in shared storage\n")
    if local_copy_unpack_count == 0 and (local_copy_count + local_wait_queue_count) > 0:
        return 1
    else:
        return 0

if __name__ == "__main__":
    try:
        result = main()
        print(result)
        sys.exit(0)
    except Exception as e:
        sys.stderr.write(f"[ERROR] {str(e)}\n")
        sys.exit(1)