# 横表

## pull-unpack

- 消息格式：1257617424/p00001_00096/1257617546_1257617585_ch109.dat.tar.zst

```sql

WITH vtable AS (
    SELECT matches[1] AS t,(matches[2]::integer)-109 AS ch,status_code
    FROM (
        SELECT regexp_matches(body, '\d{5}(\d{5})_\d{10}_ch(\d+)\.dat', 'g') matches, status_code
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

## beam-make / down-sample / fits-redist

- 消息格式：
- 1257010784/p00001_00024/t1257012766_1257012965/ch109

```sql

WITH vtable AS (
    SELECT matches[2] AS t,matches[1] AS p,(matches[3]::integer)-109 AS ch,status_code
    FROM (
        SELECT regexp_matches(body, 'p(\d{5})_\d{5}/t\d{5}(\d{5})_\d{10}/ch(\d{3})', 'g') matches, status_code
        FROM t_task
        WHERE job=1414
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

## fits-merge

- 1257010784/p00023/t1257010786_1257010965

- 消息格式：1301240224/p02955/t1301240825_1301241024

| 数据集 | 起始时间 |
| ------ | ------- |
| 1301240224 | 225 |
| 1266932744 | 2746 |
| 1266329600 | 29602 |
| 1257617424 | 17426 |
| 1255803168 | 3170 |

```sql
WITH vtable AS (
    SELECT matches[1] AS p,((matches[2]::integer)-3170)/200 AS t,status_code
    FROM (
        SELECT regexp_matches(body, 'p(\d+)/t\d{5}(\d{5})_\d{10}', 'g') matches, status_code
        FROM t_task
        WHERE job=882
    ) tt
),finished AS (
    SELECT p
    FROM (
        SELECT p,
            SUM(CASE WHEN sum_code = 0 THEN 0 ELSE 1 END) OVER (ORDER BY p) AS group_num
        FROM (
            SELECT p,
                SUM(status_code) sum_code,
                COUNT(status_code) not_null_count
            FROM vtable
            GROUP BY 1
        ) tt1
        WHERE not_null_count=24
    ) tt2
    WHERE group_num = 0
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
WHERE p NOT IN (SELECT p FROM finished)
GROUP BY 1
ORDER BY 1;
```