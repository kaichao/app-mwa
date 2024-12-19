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
            SUM(CASE WHEN sum_code = 0 THEN 0 ELSE 1 END) OVER (ORDER BY t) AS group_num
        FROM (
            SELECT t,
                SUM(status_code) sum_code,
                COUNT(status_code) not_null_count
            FROM vtable
            GROUP BY 1
        ) tt1
        WHERE not_null_count=24
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
            SUM(CASE WHEN sum_code = 0 THEN 0 ELSE 1 END) OVER (ORDER BY t,p) AS group_num
        FROM (
            SELECT t,p,
                SUM(status_code) sum_code,
                COUNT(status_code) not_null_count
            FROM vtable
            GROUP BY 1,2
        ) tt1
        WHERE not_null_count=24
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
            SUM(CASE WHEN sum_code = 0 THEN 0 ELSE 1 END) OVER (ORDER BY t,p) AS group_num
        FROM (
            SELECT t,p,
                SUM(status_code) sum_code,
                COUNT(status_code) not_null_count
            FROM vtable
            GROUP BY 1,2
        ) tt1
        WHERE not_null_count=24
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

| 数据集 | 起始时间 |
| ------ | ------- |
| 1301240224 | 225 |
| 1266932744 | 2746 |
| 1266329600 | 29602 |

```sql

WITH vtable AS (
    SELECT matches[1] AS p,((matches[2]::integer)-2746)/200 AS t,status_code
    FROM (
        SELECT regexp_matches(body, 'p(\d+)/t\d{6}(\d{4})_\d{10}', 'g') matches, status_code
        FROM t_task
        WHERE job=19
    ) tt
)
SELECT p,
    SUM(status_code) FILTER (WHERE t = 0) AS t00,
    SUM(status_code) FILTER (WHERE t = 1) AS t01,
    SUM(status_code) FILTER (WHERE t = 2) AS t02,
    SUM(status_code) FILTER (WHERE t = 3) AS t03,
    SUM(status_code) FILTER (WHERE t = 4) AS t04,
    SUM(status_code) FILTER (WHERE t = 5) AS t05,
    SUM(status_code) FILTER (WHERE t = 6) AS t06,
    SUM(status_code) FILTER (WHERE t = 7) AS t07,
    SUM(status_code) FILTER (WHERE t = 8) AS t08,
    SUM(status_code) FILTER (WHERE t = 9) AS t09,
    SUM(status_code) FILTER (WHERE t = 10) AS t10,
    SUM(status_code) FILTER (WHERE t = 11) AS t11,
    SUM(status_code) FILTER (WHERE t = 12) AS t12,
    SUM(status_code) FILTER (WHERE t = 13) AS t13,
    SUM(status_code) FILTER (WHERE t = 14) AS t14,
    SUM(status_code) FILTER (WHERE t = 15) AS t15,
    SUM(status_code) FILTER (WHERE t = 16) AS t16,
    SUM(status_code) FILTER (WHERE t = 17) AS t17,
    SUM(status_code) FILTER (WHERE t = 18) AS t18,
    SUM(status_code) FILTER (WHERE t = 19) AS t19,
    SUM(status_code) FILTER (WHERE t = 20) AS t20,
    SUM(status_code) FILTER (WHERE t = 21) AS t21,
    SUM(status_code) FILTER (WHERE t = 22) AS t22,
    SUM(status_code) FILTER (WHERE t = 23) AS t23
FROM vtable
GROUP BY 1
ORDER BY 1

```

## slot的横表
- 按计算节点的横表
```sql
WITH vtable AS (
    SELECT host, t_slot.id AS sid, t_host.ip_addr, t_job.name, serial_num, t_slot.status
    FROM t_slot 
        JOIN t_job ON(t_slot.job=t_job.id)
        JOIN t_host ON(t_slot.host=t_host.hostname)
    WHERE t_slot.host LIKE 'c-%'
        AND app=188
    ORDER BY 1,2,3
)
SELECT host, ip_addr,
    STRING_AGG( format('%s (%s)',status,sid) || '', ' ') FILTER (WHERE name = 'pull-unpack') AS pull_unpack,
    STRING_AGG(format('%s (%s)',status,sid), E'\t') FILTER (WHERE name = 'beam-maker') AS beam_maker,
    STRING_AGG(format('%s (%s)',status,sid), ' ') FILTER (WHERE name = 'down-sampler') AS down_sampler,
    STRING_AGG(format('%s (%s)',status,sid), ' ') FILTER (WHERE name = 'fits-redist') AS fits_redist,
    STRING_AGG(format('%s (%s)',status,sid), ' ') FILTER (WHERE name = 'fits-merger') AS fits_merger
FROM vtable
GROUP BY 1,2
ORDER BY 1;
```

- 系统级slot列表
```sql
WITH v_job AS (
    SELECT id, name
    FROM t_job
    WHERE app=194
        AND name in ('message-router-main','dir-list','cluster-dist','fits-24ch-push')
)
SELECT v_job.id AS jid, v_job.name AS jname, t_slot.id AS sid, host,serial_num,t_slot.status
FROM v_job JOIN t_slot ON (v_job.id=t_slot.job)
ORDER BY 1,5;
```

## extras处理的横表

```sql

WITH mapped AS (    -- label到索引号的映射
    SELECT
        name,
        ordinality AS index
    FROM 
        unnest(ARRAY[
            'before-mr', 
            'before-sema-progress-counter', 
            'after-sema-progress-counter',
            'before-sendJobRefMessage()',
            'before-leave-fromBeamMaker()',
            'before-exit'
        ]) WITH ORDINALITY AS name
),
expanded AS (
    SELECT
        id,task,jsonb_array_elements(extras->'timestamps') AS elem
    FROM t_task_exec
    WHERE task IN (
        SELECT id
        FROM t_task
        WHERE job=263 
            AND from_job='beam-maker' AND status_code=0
        ORDER BY id
        OFFSET 130000
        LIMIT 500
    )
),
numbered AS (   -- label名转换为序号
    SELECT
        task,
        elem->>'t' AS timestamp,
        index AS rn
--        ROW_NUMBER() OVER (PARTITION BY task) AS rn
    FROM expanded JOIN mapped ON(elem->>'label' = name)
),htable AS (
    SELECT task,
        MAX(CASE WHEN rn = 1 THEN timestamp END) AS t0,
        MAX(CASE WHEN rn = 2 THEN timestamp END) AS t1,
        MAX(CASE WHEN rn = 3 THEN timestamp END) AS t2,
        MAX(CASE WHEN rn = 4 THEN timestamp END) AS t3,
        MAX(CASE WHEN rn = 5 THEN timestamp END) AS t4,
        MAX(CASE WHEN rn = 6 THEN timestamp END) AS t5
    FROM numbered
    GROUP BY 1
), diff_table AS (
    SELECT
        task,
        (EXTRACT(EPOCH FROM (t1::timestamp - t0::timestamp)) * 1000000)::integer AS dt0,
        (EXTRACT(EPOCH FROM (t2::timestamp - t1::timestamp)) * 1000000)::integer AS dt1,
        (EXTRACT(EPOCH FROM (t3::timestamp - t2::timestamp)) * 1000000)::integer AS dt2,
        (EXTRACT(EPOCH FROM (t4::timestamp - t3::timestamp)) * 1000000)::integer AS dt3,
        (EXTRACT(EPOCH FROM (t5::timestamp - t4::timestamp)) * 1000000)::integer AS dt4
    FROM htable
)
SELECT t_task.begin_time, diff_table.* 
FROM diff_table JOIN t_task ON (diff_table.task=t_task.id)

```