#!/bin/bash

top_level_dir="/tmp/scalebox"
# 查找并遍历所有名为 work-* 的子目录
find "$top_level_dir" -type d -name 'work-*' | while read -r dir; do
    echo "Deleting $dir"
    rm -rf "$dir"
done

top_level_dir="/dev/shm/scalebox"
# 查找并遍历所有名为 work-* 的子目录
find "$top_level_dir" -type d -name 'work-*' | while read -r dir; do
    # 检查目录中是否包含 agent.env 文件
    if [ -f "$dir/.scalebox/agent.env" ]; then
        echo "Skipping $dir because it contains agent.env"
    else
        echo "Deleting $dir"
        rm -rf "$dir"
    fi
done


# 针对node-agent,如果目录数量大于1，保留最新的一个，删除其他的
cd ${top_level_dir}
dirs=$(ls -dt work-*/)
dir_count=$(echo "$dirs" | wc -l)
if [ "$dir_count" -gt 1 ]; then
  latest_dir=$(echo "$dirs" | head -n 1)
  dirs_to_delete=$(echo "$dirs" | tail -n +2)
  rm -rf $dirs_to_delete
  echo "Deleted directories: $dirs_to_delete"
fi

rm -f /tmp/scalebox-std*
rm -rf /tmp/scalebox/mydata/* /dev/shm/scalebox/mydata/*
