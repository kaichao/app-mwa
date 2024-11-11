
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

```sh
SOURCE_URL=scalebox@159.226.237.136:10022/raid0/tmp/mwa/tar1266932744 TARGET_URL=/data1/tmp DIR_NAME=1266932744 REGEX_FILTER= scalebox app create

SOURCE_URL=scalebox@159.226.237.136:10022/raid0/tmp/mwa TARGET_URL=/data2/tmp DIR_NAME=tar1206977296 REGEX_FILTER= scalebox app create

TARGET_URL=scalebox@159.226.237.136:10022/raid0/tmp/24ch SOURCE_URL=/data1/mydata/mwa/24ch DIR_NAME=1266932744-241102 REGEX_FILTER= scalebox app create

```

- 预拷贝文件到共享存储

在scalebox/dockerfiles/files/app-dir-copy目录下

```sh
TARGET_URL=cstu0036@60.245.128.14:65010/work1/cstu0036/mydata/mwa/tar SOURCE_URL=/data1/mydata/mwa/tar DIR_NAME=1301240224 REGEX_FILTER="/1301240[2-5]" scalebox app create

TARGET_URL=cstu0036@60.245.128.14:65010/work1/cstu0036/mydata/mwa/tar SOURCE_URL=/data1/mydata/mwa/tar DIR_NAME=1301240224 REGEX_FILTER='/130124(16\|17\|18\|19\|20\|21)' scalebox app create

TARGET_URL=cstu0036@60.245.128.14:65010/work1/cstu0036/mydata/mwa/tar SOURCE_URL=/data2/mydata/mwa/tar DIR_NAME=1266932744 REGEX_FILTER='/126693(2\|3[01])' scalebox app create

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
