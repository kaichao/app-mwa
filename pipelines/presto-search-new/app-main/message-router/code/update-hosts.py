#!/usr/local/bin/python3

# 功能：查询数据库中的t_host表，获取group_id为环境变量NODE_GROUP的所有的host的信息
# 然后分析其parameters字段，根据key为/tmp的value筛选，筛选条件为/tmp的值在环境变量VOLUME_LOW和VOLUME_HIGH之间。
# 最后修改被筛选出的host，根据ip_addr排序后，将其hostname修改为c-xxxx.${CLUSTER}的形式，
# 并将group_id修改为输入的参数。（该参数可以保证合法输入，因此无需额外检查）

import os
import sys
import json
import query_db

def myexecute(cmd):
    print("'%s'"%cmd)
    status = os.system(cmd)
    return status

def main():
    # 检查命令行参数
    if len(sys.argv) != 6:
        sys.stderr.write(f"Error: wrong args number in update_hosts.py\n")
        exit(1)
        # 获取命令行参数作为新的group_id
    new_prefix = sys.argv[1]
    group_num = int(sys.argv[2])
    
    # 获取环境变量NODE_GROUP
    node_group = os.getenv("NODE_GROUP") or "n"

    # 获取环境变量CLUSTER
    cluster = os.getenv("CLUSTER") or "local"
        
    # 获取环境变量VOLUME_LOW和VOLUME_HIGH
    try:
        volume_low = int(sys.argv[3])
        volume_mid = int(sys.argv[4])
        volume_high = int(sys.argv[5])
    except ValueError:
        sys.stderr.write("Error: volume_low, volume_high, and volume_mid must be integers\n")
        exit(1)

    
    # 查询数据库获取指定group_id的主机信息
    hosts = query_db.get_hosts_likely(node_group)
    if not hosts:
        sys.stderr.write(f"Error: No hosts found with group_id {node_group}\n")
        sys.exit(1)
    
    # 筛选符合条件的主机
    low_mid_hosts = []
    mid_high_hosts = []
    high_volume_hosts = []
    for hostname, ip_addr, params in hosts:
        # 解析parameters字段
        print(params)
        # 检查是否有/tmp键，并且其值在指定范围内
        if "/tmp" in params:
            tmp_value = params["/tmp"]
            try:
                tmp_value = int(tmp_value)
                if volume_low <= tmp_value < volume_mid:
                    low_mid_hosts.append((hostname, ip_addr, tmp_value))
                elif volume_mid <= tmp_value <= volume_high:
                    mid_high_hosts.append((hostname, ip_addr, tmp_value))
                elif tmp_value > volume_high:
                    high_volume_hosts.append((hostname, ip_addr, tmp_value))
            except (ValueError, TypeError):
                sys.stderr.write(f"Warning: Host {hostname}'s /tmp value is not a valid number\n")
                continue
        else:
            sys.stderr.write(f"Warning: Host {hostname}'s parameters field does not contain /tmp key\n")
            continue

    
    # 获取所有符合条件的主机的IP地址列表
    ip_list = [host[1] for host in mid_high_hosts]
    sorted_ips = sorted(ip_list)

    # 分组更新group_id和hostname
    if (group_num == -1):
        group_num = len(ip_list)
    if group_num <= 24:
        mode = "a"
    else:
        mode = "1"
    unused_ips = query_db.update_grouped_hosts(ip_list, new_prefix, cluster, group_num, mode)
    
    print(f"Successfully updated hostname and group_id for {len(ip_list) - len(unused_ips)} hosts")
    
    # 将未被处理的ip加入low组中
    for ip in unused_ips:
        low_mid_hosts.append(("", ip, None))

    # 获取环境变量REDIS_QUEUE
    redis_queue = os.getenv("REDIS_QUEUE") or "QUEUE_HOST"
    redis_host = os.getenv("REDIS_HOST", "localhost")
    redis_port = os.getenv("REDIS_PORT", "6379")
    cmd = f'redis-cli -h {redis_host} -p {redis_port} ZADD  {redis_queue} 1 "{ip_addr}:$(date +%s%3N)"'
    if low_mid_hosts:
        # 使用redis-cli命令发送消息
        for hostname, ip_addr, tmp_value in low_mid_hosts:
            try:
                # 使用redis-cli命令将消息添加到队列，优先级为1，并通过date命令获取当前时间戳
                status = myexecute(cmd)
                if status == 0:
                    print(f"Successfully added host {hostname}({ip_addr}) to Redis queue, /tmp value is {tmp_value}")
                else:
                    sys.stderr.write(f"Warning: Unable to add host {hostname}({ip_addr}) to Redis queue\n")
            except Exception as e:
                sys.stderr.write(f"Error: Failed to execute redis-cli command: {e}\n")
                
    
    # 处理/tmp值大于VOLUME_HIGH的主机
    if high_volume_hosts:
        # 使用redis-cli命令发送消息
        for hostname, ip_addr, tmp_value in high_volume_hosts:
            try:
                # 使用redis-cli命令将消息添加到队列，优先级为1，并通过date命令获取当前时间戳
                status = myexecute(cmd)
                status = myexecute(cmd)
                if status == 0:
                    print(f"Successfully added host {hostname}({ip_addr}) to Redis queue, /tmp value is {tmp_value}")
                else:
                    sys.stderr.write(f"Warning: Unable to add host {hostname}({ip_addr}) to Redis queue\n")
            except Exception as e:
                sys.stderr.write(f"Error: Failed to execute redis-cli command: {e}\n")
        
    print(f"Completed adding {len(high_volume_hosts) + len(low_mid_hosts)} hosts to Redis queue")

if __name__ == "__main__":
    main()

