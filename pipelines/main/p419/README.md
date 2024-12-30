
## 获取计算节点IP地址列表

```sql
SELECT string_agg(ip_addr, ' ') AS ip_addresses
FROM (
    SELECT ip_addr
    FROM t_host
    WHERE hostname LIKE 'c-%.p419' AND status='ON'
    ORDER BY hostname
) subquery;
```

## 修改t_task中纪录为READY

```sql
UPDATE t_task 
SET status_code=-1, 
    headers = jsonb_set(COALESCE(headers, '{}'::jsonb), '{repeatable}', '"yes"', true)
WHERE job=685 and status_code>0;
```

## 目录拷贝

- 从main到p419(pull)
```sh
SOURCE_URL=scalebox@159.226.237.136:10022/raid0/tmp/mwa TARGET_URL=/data2/tmp DIR_NAME=tar1257617424 REGEX_FILTER='zst$' scalebox app create

SOURCE_URL=scalebox@159.226.237.136:10022/raid0/tmp/mwa TARGET_URL=/data2/tmp DIR_NAME=tar1206977296 REGEX_FILTER= scalebox app create

TARGET_URL=scalebox@159.226.237.136:10022/raid0/tmp/24ch SOURCE_URL=/data1/mydata/mwa/24ch DIR_NAME=1266932744-241102 REGEX_FILTER= scalebox app create

```

- 从main到p419(push)

```sh
SOURCE_URL=/raid0/tmp/mwa TARGET_URL=scalebox@60.245.128.60:10022/data2/tmp DIR_NAME=tar1257617424 REGEX_FILTER='zst$' scalebox app create

```

- 预拷贝文件到共享存储

在scalebox/dockerfiles/files/app-dir-copy目录下

```sh
TARGET_URL=cstu0036@60.245.128.14:65010/work1/cstu0036/mydata/mwa/tar SOURCE_URL=/data1/mydata/mwa/tar DIR_NAME=1301240224 REGEX_FILTER="/1301240[2-5]" scalebox app create

TARGET_URL=cstu0036@60.245.128.14:65010/work1/cstu0036/mydata/mwa/tar SOURCE_URL=/data1/mydata/mwa/tar DIR_NAME=1301240224 REGEX_FILTER='/130124(16\|17\|18\|19\|20\|21)' scalebox app create

# 1266932744的前240个文件
TARGET_URL=cstu0036@60.245.128.14:65010/work2/cstu0036/mydata/mwa/tar SOURCE_URL=/data2/mydata/mwa/tar DIR_NAME=1266932744 REGEX_FILTER='/126693(2\|3[01])' scalebox app create

# 1266329600的前240个文件
TARGET_URL=cstu0036@60.245.128.14:65010/work2/cstu0036/tmp SOURCE_URL=/data2/mydata/mwa/tar DIR_NAME=1266329600 REGEX_FILTER='/1266329' scalebox app create

TARGET_URL=cstu0036@60.245.128.14:65010/work2/cstu0036/tmp SOURCE_URL=/data2/mydata/mwa/tar DIR_NAME=1266329600 scalebox app create

```


## /tmp目录下文件自动删除

- 配置文件：/usr/lib/tmpfiles.d/tmp.conf

- 检查清理文件的运行时间
```sh
systemctl list-timers | grep systemd-tmpfiles-clean
```

## 将分区文件系统从ext4换为xfs

- 步骤如下：
 
```sh
# 1. 备份数据
mkdir /backup
rsync -av /mnt/old_mount/ /backup/

# 2. 卸载分区
umount /dev/sda1



# 3. 格式化为 XFS
parted /dev/sda mklabel gpt
parted /dev/sda mkpart primary 0% 100%
mkfs.xfs -f /dev/sda

# 4. 挂载分区
mkdir /mnt/new_mount
mount /dev/sda1 /mnt/new_mount

# 5. 恢复数据
rsync -av /backup/ /mnt/new_mount/

# 6. 更新 /etc/fstab
echo '/dev/sda1 /mnt/new_mount xfs defaults 0 0' | sudo tee -a /etc/fstab

```

- blkid 命令可以直接显示块设备的文件系统类型和 UUID

```sh
blkid
```

## 优化服务器参数
- 查看用户的最大文件描述符数量
```sh
ulimit -n
```

- 修改最大文件描述符数量
```sh
echo "* soft nofile 262144" >> /etc/security/limits.conf
```
重新登录或重启后生效。

- 查看50051端口上的grpc连接数
```sh
netstat -an | grep :50051 | grep ESTABLISHED | wc -l

ss -ant | grep :50051 | grep ESTABLISHED | wc -l
```

## 数据库优化

- 将存储设备的调度器设置为 noop

在 SSD 或 NVMe 上，noop 能减少 CPU 资源占用，避免对顺序进行不必要的调整，因为固态硬盘本身可以高效处理随机和顺序 I/O。

```sh
echo noop | sudo tee /sys/block/<device>/queue/scheduler
```

- 分离数据和 WAL 文件
  - 将数据目录和 WAL 日志目录分配到不同的物理磁盘上，减少 I/O 竞争；
  - WAL目录在```/var/lib/postgresql/data/pg_wal```
  
- 使用 PostgreSQL 的 synchronous_commit 配置
   PostgreSQL 默认使用同步写入来保证事务持久性。可调整 synchronous_commit 参数，以增加异步写入的比例。
   
```sql
ALTER SYSTEM SET synchronous_commit = 'off';
```

## 进程操作

- 查看进程
```sh
ps -uxef

```

- 杀死进程
```sh

```

## 易出错节点列表

```
h01r2n19

```

## 时间戳相关

```
24-10-08T22:58:36.963924
```

```sh
date +"%y-%m-%dT%H:%M:%S.%6N"
```

- 将字符串转为postgresql的timestamp类型
```sql
SELECT TO_TIMESTAMP('24-10-08T22:58:36.963924', 'YY-MM-DD"T"HH24:MI:SS.US');
```

## 设置hostname

- centos 9
```sh
hostnamectl set-hostname scalebox-server
```

- /etc/hosts文件
```

```

## 手工修改retry

```sql
UPDATE t_task 
SET status_code=-1, headers=jsonb_set(COALESCE(headers, '{}'::jsonb), '{repeatable}', '"yes"', true)
WHERE job=230 AND status_code=137;
```

## 按job统计slot/host效率

```sql

WITH task_exec AS (
  SELECT task, slot, t2, t3
  FROM t_task_exec
  WHERE t3 IS NOT NULL AND job=584
), exec_by_slot AS (
  SELECT slot, sum(t3-t2), avg(t3-t2),max(t3),min(t2)
  FROM task_exec
  GROUP BY 1
), stat_exec_by_slot AS (
  SELECT t_slot.host, t_slot.serial_num AS num, exec_by_slot.*, max-min AS duration,
    ROUND(extract(epoch FROM exec_by_slot.sum) / 
      extract(epoch FROM exec_by_slot.max - exec_by_slot.min), 6) AS ratio
  FROM exec_by_slot JOIN t_slot ON (exec_by_slot.slot=t_slot.id) 
), exec_by_host AS (
  SELECT host, sum(sum), avg(avg),max(max),min(min),max(max)-min(min) as duration
  FROM exec_by_slot JOIN t_slot ON (exec_by_slot.slot=t_slot.id)
  GROUP BY 1
), stat_exec_by_host AS (
  SELECT *,
    ROUND(extract(epoch FROM sum) / 
      extract(epoch FROM max - min), 6) AS ratio
  FROM exec_by_host 
) 
SELECT *
FROM stat_exec_by_host
ORDER BY 1,2

```