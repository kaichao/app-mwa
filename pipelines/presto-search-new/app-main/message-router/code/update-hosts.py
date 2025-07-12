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
    if len(sys.argv) != 5:
        sys.stderr.write(f"Error: wrong args number in update_hosts.py\n")
        exit(1)
    node_group = sys.argv[1]

    # 获取环境变量CLUSTER
    cluster = os.getenv("CLUSTER") or "local"
    try:
        volume_low = int(sys.argv[2])
        volume_mid = int(sys.argv[3])
        volume_high = int(sys.argv[4])
    except ValueError:
        sys.stderr.write("Error: volume_low, volume_high, and volume_mid must be integers\n")
        exit(1)

    init_slots = int(os.environ.get('INIT_SLOTS', 0))
    print(init_slots)
    
    # 查询数据库获取指定group_id的主机信息
    hosts = query_db.get_hosts_likely(node_group)
    if not hosts:
        sys.stderr.write(f"Error: No hosts found with group_id {node_group}\n")
        sys.exit(1)
    
    # 筛选符合条件的主机
    low_hosts = []
    mid_hosts = []
    high_hosts = []
    for id, hostname, ip_addr, params in hosts:
        # 解析parameters字段
        print(params)
        # 检查是否有/tmp键，并且其值在指定范围内
        if "/tmp" in params:
            tmp_value = params["/tmp"]
            try:
                tmp_value = int(tmp_value)
                if tmp_value < volume_low:
                    low_hosts.append((hostname, ip_addr, tmp_value))
                elif volume_low <= tmp_value < volume_high:
                    mid_hosts.append((hostname, ip_addr, tmp_value))
                elif tmp_value > volume_high:
                    high_hosts.append((hostname, ip_addr, tmp_value))
            except (ValueError, TypeError):
                sys.stderr.write(f"Warning: Host {hostname}'s /tmp value is not a valid number\n")
                continue
        else:
            sys.stderr.write(f"Warning: Host {hostname}'s parameters field does not contain /tmp key\n")
            continue

    # 获取环境变量REDIS_QUEUE
    redis_queue = os.getenv("REDIS_QUEUE") or "QUEUE_HOST"
    redis_host = os.getenv("REDIS_HOST", "localhost")
    redis_port = os.getenv("REDIS_PORT", "6379")

    def send_redis_messages(ip_addr, num):
        # 使用redis-cli命令将消息添加到队列，优先级为1，并通过date命令获取当前时间戳
        cmd = f'redis-cli -h {redis_host} -p {redis_port} ZADD {redis_queue} 1 "{ip_addr}:$(date +%s%3N)"'
        for i in range(num):
            status = myexecute(cmd)
            if status != 0:
                return status
        return 0

    print(len(low_hosts), len(mid_hosts), len(high_hosts))

    if low_hosts:
        # 使用redis-cli命令发送消息
        local_slots = max(2 - init_slots, 0)
        for hostname, ip, tmp_value in low_hosts:
            try:
                # 使用redis-cli命令将消息添加到队列，优先级为1，并通过date命令获取当前时间戳
                
                status = send_redis_messages(ip, local_slots)
                if status == 0:
                    print(f"Successfully added host {hostname}({ip}) to Redis queue, /tmp value is {tmp_value}")
                else:
                    sys.stderr.write(f"Warning: Unable to add host {hostname}({ip}) to Redis queue\n")
            except Exception as e:
                sys.stderr.write(f"Error: Failed to execute redis-cli command: {e}\n")
            # print(hostname, local_slots)

    if mid_hosts:
        # 使用redis-cli命令发送消息
        local_slots = max(3 - init_slots, 0)
        for hostname, ip, tmp_value in mid_hosts:
            try:
                # 使用redis-cli命令将消息添加到队列，优先级为1，并通过date命令获取当前时间戳        
                status = send_redis_messages(ip, local_slots)
                if status == 0:
                    print(f"Successfully added host {hostname}({ip}) to Redis queue, /tmp value is {tmp_value}")
                else:
                    sys.stderr.write(f"Warning: Unable to add host {hostname}({ip}) to Redis queue\n")
            except Exception as e:
                sys.stderr.write(f"Error: Failed to execute redis-cli command: {e}\n")
            # print(hostname, local_slots)
                
    
    # 处理/tmp值大于VOLUME_HIGH的主机
    if high_hosts:
        local_slots = max(4 - init_slots, 0)
        # 使用redis-cli命令发送消息
        for hostname, ip, tmp_value in high_hosts:
            try:
                # 使用redis-cli命令将消息添加到队列，优先级为1，并通过date命令获取当前时间戳
                status = send_redis_messages(ip, local_slots)
                if status == 0:
                    print(f"Successfully added host {hostname}({ip}) to Redis queue, /tmp value is {tmp_value}")
                else:
                    sys.stderr.write(f"Warning: Unable to add host {hostname}({ip}) to Redis queue\n")
            except Exception as e:
                sys.stderr.write(f"Error: Failed to execute redis-cli command: {e}\n")
            # print(hostname, local_slots)
        

if __name__ == "__main__":
    main()

