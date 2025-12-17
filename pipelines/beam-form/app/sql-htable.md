# 横表

## pull-unpack进度

- 消息格式：1257617424/p00001_00096/1257617546_1257617585_ch109.dat.tar.zst

```sql

WITH vtable AS (
    SELECT matches[2] AS t,(matches[3]::integer)-109 AS ch,status_code
    FROM (
        SELECT regexp_matches(body, 'p(\d{5})_\d{5}/\d{5}(\d{5})_\d{10}_ch(\d+)\.dat', 'g') matches, status_code
        FROM t_task
        WHERE mod_id=220
    ) tt0
--    WHERE (matches[1]::integer) between 4441 and 4800
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

## pull-unpack进度（tablefunc实现，INTERLEAVED_DAT）

```sql

WITH htable AS (
  SELECT *
  FROM crosstab (
    $$
    WITH vtable AS (
      SELECT body_s[2] AS t,(to_host_s[1]::integer) AS ch,status_code
      FROM (
        SELECT regexp_matches(body, 'p(\d{5})_\d{5}/\d{5}(\d{5})_\d{10}_ch(\d+)\.dat', 'g') body_s, 
            regexp_matches(to_host, '-\d{2}(\d{2})\.', 'g') to_host_s, 
            status_code
        FROM t_task
        WHERE mod_id=471
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
    SELECT t, ch, status_code
    FROM vtable
    WHERE t NOT IN (SELECT t FROM finished)
    ORDER BY t, ch
    $$,
    $$ SELECT generate_series(0, 23) $$
  ) AS ct (
    t text,
    c00 varchar, c01 varchar, c02 varchar, c03 varchar, c04 varchar, c05 varchar, 
    c06 varchar, c07 varchar, c08 varchar, c09 varchar, c10 varchar, c11 varchar, 
    c12 varchar, c13 varchar, c14 varchar, c15 varchar, c16 varchar, c17 varchar, 
    c18 varchar, c19 varchar, c20 varchar, c21 varchar, c22 varchar, c23 varchar
  )
  ORDER BY t
)
SELECT *
FROM htable;

```


## pull-unpack任务分布（tablefunc实现）

```sql
WITH htable AS (
  SELECT *
  FROM crosstab (
    $$
    WITH vtable AS (
      SELECT 
        matches[2] AS t, 
        (matches[3]::integer - 109) AS ch, 
        substring(to_host FROM 'c-(\d{4})\.p419') AS to_host
      FROM t_task,
        regexp_matches(body, 'p(\d{5})_\d{5}/\d{5}(\d{5})_\d{10}_ch(\d+)\.dat', 'g') AS matches
      WHERE mod_id = 1959
    )
    SELECT t, ch, to_host
    FROM vtable
    ORDER BY t, ch
    $$,
    $$ SELECT generate_series(0, 23) $$
  ) AS ct (
    t text,
    c00 varchar, c01 varchar, c02 varchar, c03 varchar, c04 varchar, c05 varchar, 
    c06 varchar, c07 varchar, c08 varchar, c09 varchar, c10 varchar, c11 varchar, 
    c12 varchar, c13 varchar, c14 varchar, c15 varchar, c16 varchar, c17 varchar, 
    c18 varchar, c19 varchar, c20 varchar, c21 varchar, c22 varchar, c23 varchar
  )
  ORDER BY t
), numbered AS (
  SELECT *, ROW_NUMBER() OVER (ORDER BY t) AS rn
  FROM htable
)
SELECT *
FROM numbered
WHERE (rn - 1) % 5 = 0
ORDER BY rn;

```

通过numbered，实现每5行输出1条横表纪录.

## beam-make / down-sample / fits-redist 进度

- 消息格式：
- 1257010784/p00001_00024/t1257012766_1257012965/ch109

```sql

WITH vtable AS (
    SELECT body_s[2] AS t,body_s[1] AS p,(to_host_s[1]::integer) AS ch,status_code
    FROM (
        SELECT regexp_matches(body, 'p(\d{5})_\d{5}/t\d{5}(\d{5})_\d{10}/ch(\d{3})', 'g') body_s, 
        regexp_matches(headers->>'to_host', '\d{2}-(\d+)\.', 'g') to_host_s, 
        status_code
        FROM t_task
        WHERE mod_id=17
    ) tt
--    WHERE (matches[1]::integer) between 4441 and 4800
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

## beam-make / down-sample / fits-redist (INTERLEAVED_DAT) 进度
异常排查时，用于定位出错slot。

- 消息格式：1257010784/p00001_00024/t1257012766_1257012965/ch109
- to_host : c-0023.p419

针对前述sql语句，将

```sql
WITH vtable AS (
    SELECT matches[2] AS t,matches[1] AS p,(matches[3]::integer)-109 AS ch,status_code
    FROM (
        SELECT regexp_matches(body, 'p(\d{5})_\d{5}/t\d{5}(\d{5})_\d{10}/ch(\d{3})', 'g') matches, status_code
        FROM t_task
        WHERE mod_id=221
    ) tt
--    WHERE (matches[1]::integer) between 4441 and 4800
) ...
```

改为

```sql

WITH vtable AS (
    SELECT body_s[2] AS t,body_s[1] AS p,(to_host_s[1]::integer) AS ch,status_code
    FROM (
        SELECT regexp_matches(body, 'p(\d{5})_\d{5}/t\d{5}(\d{5})_\d{10}/ch(\d{3})', 'g') body_s, 
        regexp_matches(to_host, '-\d{2}(\d{2})\.', 'g') to_host_s, 
        status_code
        FROM t_task
        WHERE mod_id=354
    ) tt
--    WHERE (body_s[1]::integer) between 4441 and 4800
) ...
```

## beam-make / down-sample / fits-redist (INTERLEAVED_DAT+POINTING_FIRST) 进度
正常运行使用。

- 消息格式：1257010784/p00001_00024/t1257012766_1257012965/ch109
- to_host : c-0023.p419

针对前述sql语句，将

```sql
SELECT t, p,
    sum(CASE ch WHEN 0 THEN status_code END) AS c00,
    ...
```

改为：
```sql
SELECT p, t,
    sum(CASE ch WHEN 0 THEN status_code END) AS c00,
    ...
```

## fits-merge 进度

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
    SELECT matches[1] AS p,((matches[2]::integer)-3170)/160 AS t,status_code
    FROM (
        SELECT regexp_matches(body, 'p(\d+)/t\d{5}(\d{5})_\d{10}', 'g') matches, status_code
        FROM t_task
        WHERE mod_id=243
    ) tt
--    WHERE (matches[1]::integer) between 4441 and 4800
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


```sql

WITH htable AS (
SELECT *
FROM crosstab(
    -- 将 vtable 和 finished 的逻辑嵌入到 crosstab 的源查询中
    '
    WITH vtable AS (
        SELECT matches[1] AS p, ((matches[2]::integer) - 6649) / 200 AS t, status_code
        FROM (
            SELECT regexp_matches(body, ''p(\d+)/t\d{5}(\d{5})_\d{10}'', ''g'') matches, status_code
            FROM t_task
            WHERE mod_id = 981
        ) tt
--      WHERE (matches[1]::integer) between 4441 and 4800
    ), finished AS (
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
            WHERE not_null_count = 24
        ) tt2
        WHERE group_num = 0
    )
    SELECT p, t, SUM(status_code) AS status_sum
    FROM vtable
    WHERE p NOT IN (SELECT p FROM finished)
    GROUP BY p, t
    ORDER BY p, t
    ',
    -- 类别查询保持不变
    'SELECT generate_series(0, 29) AS t'
) AS ct (
    p text,
      t00 integer, t01 integer, t02 integer, t03 integer, t04 integer, t05 integer
    , t06 integer, t07 integer, t08 integer, t09 integer, t10 integer, t11 integer
    , t12 integer, t13 integer, t14 integer, t15 integer, t16 integer, t17 integer
    , t18 integer, t19 integer, t20 integer, t21 integer, t22 integer, t23 integer
    , t24 integer, t25 integer, t26 integer, t27 integer, t28 integer, t29 integer 
)
ORDER BY p
)
SELECT *
FROM htable;

```



## beam-make状态

### progress横表

```sql

WITH htable AS (
  SELECT *
  FROM crosstab (
    $$
    WITH vtable AS (
      SELECT name_s[1] AS g, name_s[2]::integer AS ch, value
      FROM (
        SELECT regexp_matches(name, '-(\d{2})(\d{2})', 'g') name_s, 
            value
        FROM t_semaphore
        WHERE name LIKE 'task_progress:beam-make:c%' AND
            app=103
      ) tt0
    )
    SELECT g, ch, value
    FROM vtable
    ORDER BY g, ch
    $$,
    $$ SELECT generate_series(0, 23) $$
  ) AS ct (
    t text,
    c00 integer, c01 integer, c02 integer, c03 integer, c04 integer, c05 integer, 
    c06 integer, c07 integer, c08 integer, c09 integer, c10 integer, c11 integer, 
    c12 integer, c13 integer, c14 integer, c15 integer, c16 integer, c17 integer, 
    c18 integer, c19 integer, c20 integer, c21 integer, c22 integer, c23 integer
  )
  ORDER BY t
)
SELECT *
FROM htable

```

### slot状态

```sql

WITH htable AS (
  SELECT *
  FROM crosstab (
    $$
    WITH vtable AS (
      SELECT host_s[1] AS g, host_s[2]::integer AS h, 
      STRING_AGG ( st, ' ' ORDER BY seq) s
      FROM (
        SELECT regexp_matches(host, '-(\d{2})(\d{2})', 'g') host_s, 
            seq, 
            CASE status
                WHEN 'ON' THEN 'O'
                WHEN 'READY' THEN 'R'
                WHEN 'ERROR' THEN 'E'
                ELSE status -- 如果有其他值，保留原值
            END AS st
        FROM t_slot
        WHERE mod_id=608
        ORDER BY host, seq
      ) tt0
      GROUP BY 1,2
    )
    SELECT g, h, s
    FROM vtable
    ORDER BY g, h
    $$,
    $$ SELECT generate_series(0, 23) $$
  ) AS ct (
    t text,
    c00 varchar, c01 varchar, c02 varchar, c03 varchar, c04 varchar, c05 varchar, 
    c06 varchar, c07 varchar, c08 varchar, c09 varchar, c10 varchar, c11 varchar, 
    c12 varchar, c13 varchar, c14 varchar, c15 varchar, c16 varchar, c17 varchar, 
    c18 varchar, c19 varchar, c20 varchar, c21 varchar, c22 varchar, c23 varchar
  )
  ORDER BY t
)
SELECT *
FROM htable

```
