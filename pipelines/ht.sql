-- 打包文件的横表
-- 1301240224/1301240405_1301240434_ch114.dat.tar.zst~b00	
WITH ht AS (
    SELECT matches[1] AS t,(matches[3]::integer)-109 AS ch,status_code
    FROM (
        SELECT regexp_matches(key_message, '(\d{10})_(\d{10})_ch(\d+)\.dat', 'g') matches, status_code
        FROM t_task
        WHERE job=1141
    ) tt
)
SELECT t,
   sum(CASE ch WHEN 0 THEN status_code END) AS f00,
   sum(CASE ch WHEN 1 THEN status_code END) AS f01,
   sum(CASE ch WHEN 2 THEN status_code END) AS f02,
   sum(CASE ch WHEN 3 THEN status_code END) AS f03,
   sum(CASE ch WHEN 4 THEN status_code END) AS f03,
   sum(CASE ch WHEN 5 THEN status_code END) AS f05,
   sum(CASE ch WHEN 6 THEN status_code END) AS f06,
   sum(CASE ch WHEN 7 THEN status_code END) AS f07,
   sum(CASE ch WHEN 8 THEN status_code END) AS f08,
   sum(CASE ch WHEN 9 THEN status_code END) AS f09,
   sum(CASE ch WHEN 10 THEN status_code END) AS f10,
   sum(CASE ch WHEN 11 THEN status_code END) AS f11,
   sum(CASE ch WHEN 12 THEN status_code END) AS f12,
   sum(CASE ch WHEN 13 THEN status_code END) AS f13,
   sum(CASE ch WHEN 14 THEN status_code END) AS f14,
   sum(CASE ch WHEN 15 THEN status_code END) AS f15,
   sum(CASE ch WHEN 16 THEN status_code END) AS f16,
   sum(CASE ch WHEN 17 THEN status_code END) AS f17,
   sum(CASE ch WHEN 18 THEN status_code END) AS f18,
   sum(CASE ch WHEN 19 THEN status_code END) AS f19,
   sum(CASE ch WHEN 20 THEN status_code END) AS f20,
   sum(CASE ch WHEN 21 THEN status_code END) AS f21,
   sum(CASE ch WHEN 22 THEN status_code END) AS f22,
   sum(CASE ch WHEN 23 THEN status_code END) AS f23
FROM ht
GROUP BY 1
ORDER BY 1
;


-- beam-former的横表

-- 1301240224/1301240225_1301240374/110/00385_00408

WITH ht AS (
    SELECT matches[1] AS t,matches[4] AS p,(matches[3]::integer)-109 AS ch,status_code
    FROM (
        SELECT regexp_matches(key_message, '(\d{10})_(\d{10})/(\d+)/(\d+_\d+)', 'g') matches, status_code
        FROM t_task
        WHERE job=1253
    ) tt
)
SELECT t, p,
   sum(CASE ch WHEN 0 THEN status_code END) AS f00,
   sum(CASE ch WHEN 1 THEN status_code END) AS f01,
   sum(CASE ch WHEN 2 THEN status_code END) AS f02,
   sum(CASE ch WHEN 3 THEN status_code END) AS f03,
   sum(CASE ch WHEN 4 THEN status_code END) AS f03,
   sum(CASE ch WHEN 5 THEN status_code END) AS f05,
   sum(CASE ch WHEN 6 THEN status_code END) AS f06,
   sum(CASE ch WHEN 7 THEN status_code END) AS f07,
   sum(CASE ch WHEN 8 THEN status_code END) AS f08,
   sum(CASE ch WHEN 9 THEN status_code END) AS f09,
   sum(CASE ch WHEN 10 THEN status_code END) AS f10,
   sum(CASE ch WHEN 11 THEN status_code END) AS f11,
   sum(CASE ch WHEN 12 THEN status_code END) AS f12,
   sum(CASE ch WHEN 13 THEN status_code END) AS f13,
   sum(CASE ch WHEN 14 THEN status_code END) AS f14,
   sum(CASE ch WHEN 15 THEN status_code END) AS f15,
   sum(CASE ch WHEN 16 THEN status_code END) AS f16,
   sum(CASE ch WHEN 17 THEN status_code END) AS f17,
   sum(CASE ch WHEN 18 THEN status_code END) AS f18,
   sum(CASE ch WHEN 19 THEN status_code END) AS f19,
   sum(CASE ch WHEN 20 THEN status_code END) AS f20,
   sum(CASE ch WHEN 21 THEN status_code END) AS f21,
   sum(CASE ch WHEN 22 THEN status_code END) AS f22,
   sum(CASE ch WHEN 23 THEN status_code END) AS f23
FROM ht
GROUP BY 1,2
ORDER BY 1,2
;
