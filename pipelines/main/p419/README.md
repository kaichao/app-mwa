
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

## cluster-dist/pull-unpack任务的横表
- 消息格式：1301240224/1301240405_1301240434_ch114.dat.tar.zst~b00	

```sql

WITH vtable AS (
    SELECT matches[1] AS t,(matches[2]::integer)-109 AS ch,status_code
    FROM (
        SELECT regexp_matches(body, '\d{6}(\d{4})_\d{10}_ch(\d+)\.dat', 'g') matches, status_code
        FROM t_task
        WHERE job=11
    ) tt0
),finished AS (
    SELECT t
    FROM (
        SELECT t,
            SUM(CASE WHEN count = 0 THEN 0 ELSE 1 END) OVER (ORDER BY t) AS group_num
        FROM (
            SELECT t,sum(status_code) count
            FROM vtable
            GROUP BY 1
        ) tt1
    ) tt2 
    WHERE group_num = 0
)
SELECT t,
    sum(CASE ch WHEN 0 THEN status_code END) AS c00,
    sum(CASE ch WHEN 1 THEN status_code END) AS c01,
    sum(CASE ch WHEN 2 THEN status_code END) AS c02,
    sum(CASE ch WHEN 3 THEN status_code END) AS c03,
    sum(CASE ch WHEN 4 THEN status_code END) AS c04,
    sum(CASE ch WHEN 5 THEN status_code END) AS c05,
    sum(CASE ch WHEN 6 THEN status_code END) AS c06,
    sum(CASE ch WHEN 7 THEN status_code END) AS c07,
    sum(CASE ch WHEN 8 THEN status_code END) AS c08,
    sum(CASE ch WHEN 9 THEN status_code END) AS c09,
    sum(CASE ch WHEN 10 THEN status_code END) AS c10,
    sum(CASE ch WHEN 11 THEN status_code END) AS c11,
    sum(CASE ch WHEN 12 THEN status_code END) AS c12,
    sum(CASE ch WHEN 13 THEN status_code END) AS c13,
    sum(CASE ch WHEN 14 THEN status_code END) AS c14,
    sum(CASE ch WHEN 15 THEN status_code END) AS c15,
    sum(CASE ch WHEN 16 THEN status_code END) AS c16,
    sum(CASE ch WHEN 17 THEN status_code END) AS c17,
    sum(CASE ch WHEN 18 THEN status_code END) AS c18,
    sum(CASE ch WHEN 19 THEN status_code END) AS c19,
    sum(CASE ch WHEN 20 THEN status_code END) AS c20,
    sum(CASE ch WHEN 21 THEN status_code END) AS c21,
    sum(CASE ch WHEN 22 THEN status_code END) AS c22,
    sum(CASE ch WHEN 23 THEN status_code END) AS c23
FROM vtable
WHERE t NOT IN (SELECT t FROM finished)
GROUP BY 1
ORDER BY 1

```
其中，finished为以完成，并且所有任务返回码都为0的时间列表
- tt1：统计出某时间点上，所有返回码的加和；
- tt2：按时间 () 排序，用窗口函数SUM对count列的非零值进行累加。最初的count为 0 的记录会得到相同的group_num值（为0），一旦遇到 count 非零的记录，group_num 值将开始递增；
- finished：过滤 group_num 不为 0 的记录。

## beam-former任务的横表

- 消息格式：1301240224/1301240225_1301240374/110/00385_00408

```sql

WITH vtable AS (
    SELECT matches[1] AS t,matches[3] AS p,(matches[2]::integer)-109 AS ch,status_code
    FROM (
        SELECT regexp_matches(body, '\d{6}(\d{4})_\d{10}/(\d{3})/(\d{5})_\d{5}', 'g') matches, status_code
        FROM t_task
        WHERE job=12
    ) tt
),finished AS (
    SELECT t,p
    FROM (
        SELECT t,p,
            SUM(CASE WHEN count = 0 THEN 0 ELSE 1 END) OVER (ORDER BY t,p) AS group_num
        FROM (
            SELECT t,p,sum(status_code) count
            FROM vtable
            GROUP BY 1,2
        ) tt1
    ) tt2 
    WHERE group_num = 0
)
SELECT t, p,
    sum(CASE ch WHEN 0 THEN status_code END) AS c00,
    sum(CASE ch WHEN 1 THEN status_code END) AS c01,
    sum(CASE ch WHEN 2 THEN status_code END) AS c02,
    sum(CASE ch WHEN 3 THEN status_code END) AS c03,
    sum(CASE ch WHEN 4 THEN status_code END) AS c04,
    sum(CASE ch WHEN 5 THEN status_code END) AS c05,
    sum(CASE ch WHEN 6 THEN status_code END) AS c06,
    sum(CASE ch WHEN 7 THEN status_code END) AS c07,
    sum(CASE ch WHEN 8 THEN status_code END) AS c08,
    sum(CASE ch WHEN 9 THEN status_code END) AS c09,
    sum(CASE ch WHEN 10 THEN status_code END) AS c10,
    sum(CASE ch WHEN 11 THEN status_code END) AS c11,
    sum(CASE ch WHEN 12 THEN status_code END) AS c12,
    sum(CASE ch WHEN 13 THEN status_code END) AS c13,
    sum(CASE ch WHEN 14 THEN status_code END) AS c14,
    sum(CASE ch WHEN 15 THEN status_code END) AS c15,
    sum(CASE ch WHEN 16 THEN status_code END) AS c16,
    sum(CASE ch WHEN 17 THEN status_code END) AS c17,
    sum(CASE ch WHEN 18 THEN status_code END) AS c18,
    sum(CASE ch WHEN 19 THEN status_code END) AS c19,
    sum(CASE ch WHEN 20 THEN status_code END) AS c20,
    sum(CASE ch WHEN 21 THEN status_code END) AS c21,
    sum(CASE ch WHEN 22 THEN status_code END) AS c22,
    sum(CASE ch WHEN 23 THEN status_code END) AS c23
FROM vtable
WHERE (t,p) NOT IN (SELECT t,p FROM finished)
GROUP BY 1,2
ORDER BY 1,2

```

## down-sampler任务的横表

- 消息格式：1301240224/p02218/t1301240425_1301240624/ch128.fits

```sql

WITH vtable AS (
    SELECT matches[2] AS t,matches[1] AS p,(matches[3]::integer)-109 AS ch,status_code
    FROM (
        SELECT regexp_matches(body, 'p(\d+)/t\d{6}(\d{4})_\d{10}/ch(\d{3})', 'g') matches, status_code
        FROM t_task
        WHERE job=13
    ) tt
),finished AS (
    SELECT t,p
    FROM (
        SELECT t,p,
            SUM(CASE WHEN count = 0 THEN 0 ELSE 1 END) OVER (ORDER BY t,p) AS group_num
        FROM (
            SELECT t,p,sum(status_code) count
            FROM vtable
            GROUP BY 1,2
        ) tt1
    ) tt2 
    WHERE group_num = 0
)
SELECT t, p,
    sum(CASE ch WHEN 0 THEN status_code END) AS c00,
    sum(CASE ch WHEN 1 THEN status_code END) AS c01,
    sum(CASE ch WHEN 2 THEN status_code END) AS c02,
    sum(CASE ch WHEN 3 THEN status_code END) AS c03,
    sum(CASE ch WHEN 4 THEN status_code END) AS c04,
    sum(CASE ch WHEN 5 THEN status_code END) AS c05,
    sum(CASE ch WHEN 6 THEN status_code END) AS c06,
    sum(CASE ch WHEN 7 THEN status_code END) AS c07,
    sum(CASE ch WHEN 8 THEN status_code END) AS c08,
    sum(CASE ch WHEN 9 THEN status_code END) AS c09,
    sum(CASE ch WHEN 10 THEN status_code END) AS c10,
    sum(CASE ch WHEN 11 THEN status_code END) AS c11,
    sum(CASE ch WHEN 12 THEN status_code END) AS c12,
    sum(CASE ch WHEN 13 THEN status_code END) AS c13,
    sum(CASE ch WHEN 14 THEN status_code END) AS c14,
    sum(CASE ch WHEN 15 THEN status_code END) AS c15,
    sum(CASE ch WHEN 16 THEN status_code END) AS c16,
    sum(CASE ch WHEN 17 THEN status_code END) AS c17,
    sum(CASE ch WHEN 18 THEN status_code END) AS c18,
    sum(CASE ch WHEN 19 THEN status_code END) AS c19,
    sum(CASE ch WHEN 20 THEN status_code END) AS c20,
    sum(CASE ch WHEN 21 THEN status_code END) AS c21,
    sum(CASE ch WHEN 22 THEN status_code END) AS c22,
    sum(CASE ch WHEN 23 THEN status_code END) AS c23
FROM vtable
WHERE (t,p) NOT IN (SELECT t,p FROM finished)
GROUP BY 1,2
ORDER BY 1,2

```

## fits-redist任务的横表

- 消息格式：1301240224/p02218/t1301240425_1301240624/ch128.fits

- 每行仅有23个非空数值。

```sql

WITH vtable AS (
    SELECT matches[2] AS t,matches[1] AS p,(matches[3]::integer)-109 AS ch,status_code
    FROM (
        SELECT regexp_matches(body, 'p(\d+)/t\d{6}(\d{4})_\d{10}/ch(\d{3})', 'g') matches, status_code
        FROM t_task
        WHERE job=26
    ) tt
),finished AS (
    SELECT t,p
    FROM (
        SELECT t,p,
            SUM(CASE WHEN sum_code = 0 THEN 0 ELSE 1 END) OVER (ORDER BY t,p) AS group_num
        FROM (
            SELECT t,p,
                sum(status_code) sum_code,
                COUNT(status_code) not_null_count
            FROM vtable
            GROUP BY 1,2
        ) tt1
        WHERE not_null_count=23
    ) tt2 
    WHERE group_num = 0
)
SELECT t, p,
    sum(CASE ch WHEN 0 THEN status_code END) AS c00,
    sum(CASE ch WHEN 1 THEN status_code END) AS c01,
    sum(CASE ch WHEN 2 THEN status_code END) AS c02,
    sum(CASE ch WHEN 3 THEN status_code END) AS c03,
    sum(CASE ch WHEN 4 THEN status_code END) AS c04,
    sum(CASE ch WHEN 5 THEN status_code END) AS c05,
    sum(CASE ch WHEN 6 THEN status_code END) AS c06,
    sum(CASE ch WHEN 7 THEN status_code END) AS c07,
    sum(CASE ch WHEN 8 THEN status_code END) AS c08,
    sum(CASE ch WHEN 9 THEN status_code END) AS c09,
    sum(CASE ch WHEN 10 THEN status_code END) AS c10,
    sum(CASE ch WHEN 11 THEN status_code END) AS c11,
    sum(CASE ch WHEN 12 THEN status_code END) AS c12,
    sum(CASE ch WHEN 13 THEN status_code END) AS c13,
    sum(CASE ch WHEN 14 THEN status_code END) AS c14,
    sum(CASE ch WHEN 15 THEN status_code END) AS c15,
    sum(CASE ch WHEN 16 THEN status_code END) AS c16,
    sum(CASE ch WHEN 17 THEN status_code END) AS c17,
    sum(CASE ch WHEN 18 THEN status_code END) AS c18,
    sum(CASE ch WHEN 19 THEN status_code END) AS c19,
    sum(CASE ch WHEN 20 THEN status_code END) AS c20,
    sum(CASE ch WHEN 21 THEN status_code END) AS c21,
    sum(CASE ch WHEN 22 THEN status_code END) AS c22,
    SUM(status_code) FILTER (WHERE ch = 23) AS c23
FROM vtable
WHERE (t,p) NOT IN (SELECT t,p FROM finished)
GROUP BY 1,2
ORDER BY 1,2
```

## fits-merger任务的横表

- 消息格式：1301240224/p02955/t1301240825_1301241024

```sql

WITH vtable AS (
    SELECT matches[1] AS p,matches[2] AS t,status_code
    FROM (
        SELECT regexp_matches(body, 'p(\d+)/t\d{6}(\d{4})_\d{10}', 'g') matches, status_code
        FROM t_task
        WHERE job=19
    ) tt
)
SELECT p,
    SUM(status_code) FILTER (WHERE t = '0225') AS c00,
    SUM(status_code) FILTER (WHERE t = '0425') AS c01,
    SUM(status_code) FILTER (WHERE t = '0625') AS c02,
    SUM(status_code) FILTER (WHERE t = '0825') AS c03,
    SUM(status_code) FILTER (WHERE t = '1025') AS c04,
    SUM(status_code) FILTER (WHERE t = '1225') AS c05,
    SUM(status_code) FILTER (WHERE t = '1425') AS c06,
    SUM(status_code) FILTER (WHERE t = '1625') AS c07,
    SUM(status_code) FILTER (WHERE t = '1825') AS c08,
    SUM(status_code) FILTER (WHERE t = '2025') AS c09,
    SUM(status_code) FILTER (WHERE t = '2225') AS c10,
    SUM(status_code) FILTER (WHERE t = '2425') AS c11,
    SUM(status_code) FILTER (WHERE t = '2625') AS c12,
    SUM(status_code) FILTER (WHERE t = '2825') AS c13,
    SUM(status_code) FILTER (WHERE t = '3025') AS c14,
    SUM(status_code) FILTER (WHERE t = '3225') AS c15,
    SUM(status_code) FILTER (WHERE t = '3425') AS c16,
    SUM(status_code) FILTER (WHERE t = '3625') AS c17,
    SUM(status_code) FILTER (WHERE t = '3825') AS c18,
    SUM(status_code) FILTER (WHERE t = '4025') AS c19,
    SUM(status_code) FILTER (WHERE t = '4225') AS c20,
    SUM(status_code) FILTER (WHERE t = '4425') AS c21,
    SUM(status_code) FILTER (WHERE t = '4625') AS c22,
    SUM(status_code) FILTER (WHERE t = '4825') AS c23
FROM vtable
GROUP BY 1
ORDER BY 1

```

## /tmp目录下文件自动删除

- 配置文件：/usr/lib/tmpfiles.d/tmp.conf

- 检查清理文件的运行时间
```sh
systemctl list-timers | grep systemd-tmpfiles-clean
```

## 优化服务器参数

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
