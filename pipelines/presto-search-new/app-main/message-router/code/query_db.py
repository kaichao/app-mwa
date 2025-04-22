import os
import sys
import psycopg2
from psycopg2 import pool
import socket
import string

db = None

def get_db():
    global db
    if db is not None:
        return db

    # 获取数据库连接配置
    database_url = os.getenv("DATABASE_URL")
    if not database_url:
        pg_host = os.getenv("PGHOST")
        pg_port = os.getenv("PGPORT")
        if not pg_host and os.getenv("JOB_NAME"):
            # 如果在代理中，使用 GRPC_SERVER 作为默认的服务器
            grpc_server = os.getenv("GRPC_SERVER")
            pg_host = grpc_server.split(":")[0]
            sys.stderr.write(f"[INFO] Set GRPC_SERVER {grpc_server} as default db server.")
        if not pg_host:
            pg_host = os.getenv("LOCAL_ADDR")
        if not pg_host:
            # 使用本地 IP 作为默认的数据库服务器
            local_ip = get_local_ip()
            pg_host = local_ip
            sys.stderr.write(f"[INFO] Set local IP {local_ip} as default db server.")

        # 如果有端口在 PGHOST 中定义
        if ":" in pg_host:
            pg_host, pg_port = pg_host.split(":", 1)
        if not pg_port:
            pg_port = "5432"

        pg_user = os.getenv("PGUSER", "scalebox")
        pg_pass = os.getenv("PGPASS", "changeme")
        pg_db = os.getenv("PGDB", "scalebox")

        # database_url = f"postgresql://{pg_user}:{pg_pass}@{pg_host}:{pg_port}/{pg_db}"

    # 获取连接池配置
    max_idles = int(os.getenv("PG_MAX_IDLE_CONNS", "50"))
    max_opens = int(os.getenv("PG_MAX_OPEN_CONNS", "20"))

    # 设置数据库连接池
    try:
        db = psycopg2.pool.SimpleConnectionPool(1, max_opens, dbname=pg_db, user=pg_user, password=pg_pass,
                                                host=pg_host, port=pg_port, connect_timeout=500)
    except Exception as e:
        raise Exception(f"Unable to connect to database: {e}")

    # 返回连接池
    return db

def get_local_ip():
    # 示例：获取本地 IP 地址的函数实现
    hostname = socket.gethostname()
    return socket.gethostbyname(hostname)


def get_connection():
    return get_db().getconn()

def put_connection(conn):
    get_db().putconn(conn)

def close_connection():
    get_db().closeall()

# 给定一个 group_id，返回该组中所有的 host的 hostname
def get_hosts(group_id):
    conn = get_connection()
    cur = conn.cursor()
    cur.execute("SELECT hostname FROM t_host WHERE group_id = %s", (group_id,))
    rows = cur.fetchall()
    put_connection(conn)
    return rows

def get_hosts_likely(group_id):
    conn = get_connection()
    cur = conn.cursor()
    cur.execute("SELECT hostname, ip_addr, parameters FROM t_host WHERE group_id ~ %s", (group_id,))
    rows = cur.fetchall()
    put_connection(conn)
    return rows

def get_hosts_likely_ordered(rfi_id, unpack_id, dedisp_id, group_id):
    sql = """
    SELECT 
    h.hostname AS host_name,
    COALESCE(SUM(CASE WHEN t.job = %s THEN 1 ELSE 0 END), 0) AS rfi_alloc,
    COALESCE(SUM(CASE WHEN t.job = %s THEN 1 ELSE 0 END), 0) AS unpack_alloc,
    COALESCE(SUM(CASE WHEN t.job = %s THEN 1 ELSE 0 END), 0) AS dedisp_alloc
    FROM 
        t_host h
    LEFT JOIN 
        t_task t ON h.hostname = t.to_host
    AND 
        t.status_code in (-1, -2, -3)
    WHERE 
        h.group_id ~ %s
    GROUP BY 
        h.hostname
    ORDER BY 
        unpack_alloc, rfi_alloc, dedisp_alloc, host_name;
    """
    conn = get_connection()
    cur = conn.cursor()
    cur.execute(sql, (rfi_id, unpack_id, dedisp_id, group_id))
    rows = cur.fetchall()
    put_connection(conn)
    return rows


# 给定一个job_id， 返回和该job具有相同app的所有job的id和name
def get_same_app_jobs(job_id):
    conn = get_connection()
    cur = conn.cursor()
    cur.execute("SELECT id, name FROM t_job WHERE app = (SELECT app FROM t_job WHERE id = %s)", (job_id,))
    rows = cur.fetchall()
    put_connection(conn)
    return rows

# 给定一个job_id， 返回和该job具有相同app的job中，name为指定值的job的id
def get_same_app_job_by_name(job_id, name):
    conn = get_connection()
    cur = conn.cursor()
    cur.execute("SELECT id FROM t_job WHERE app = (SELECT app FROM t_job WHERE id = %s) AND name = %s", (job_id, name))
    rows = cur.fetchall()
    put_connection(conn)
    return rows


# 给定一个job_id和host，返回这个host上对应job的slot的id和status
def get_job_slot(job_id, host):
    conn = get_connection()
    cur = conn.cursor()
    cur.execute("SELECT id, status FROM t_slot WHERE job = %s AND host = %s", (job_id, host))
    rows = cur.fetchall()
    put_connection(conn)
    return rows

# 为指定job在指定host上创建若干slot，初始状态为READY
def create_job_slots(job_id, host, num_slots):
    conn = get_connection()
    cur = conn.cursor()
    for i in range(num_slots):
        cur.execute("INSERT INTO t_slot (host, serial_num, status, job) VALUES (%s, %s, %s, %s)", (host, i, "READY", job_id))
    conn.commit()
    put_connection(conn)

def get_host_by_ip(ip):
    conn = get_connection()
    cur = conn.cursor()
    cur.execute("SELECT hostname FROM t_host WHERE ip_addr = %s", (ip,))
    rows = cur.fetchall()
    put_connection(conn)
    return rows

# 更新t_host表中,ip_addr在传入的ip_list中的记录的group_id为group_id
def update_host_group_id(ip_list, group_id):
    if not ip_list:
        return  # 空列表时不做任何操作

    conn = get_connection()
    cur = conn.cursor()

    # 构建 SQL 占位符
    placeholders = ','.join(['%s'] * len(ip_list))
    sql = f"UPDATE t_host SET group_id = %s WHERE ip_addr IN ({placeholders})"
    params = [group_id] + ip_list

    cur.execute(sql, params)
    conn.commit()
    put_connection(conn)


# 更新t_host表中,ip_addr在传入的ip_list中的记录的hostname为指定格式
def update_host_hostname(ip_list, hostname_prefix, cluster):
    if not ip_list:
        return

    conn = get_connection()
    cur = conn.cursor()

    # 构建 IP → 新 hostname 映射（按 IP 排序后编号）
    sorted_ips = sorted(ip_list)
    ip_hostname_pairs = [
        (ip, f"{hostname_prefix}-{i:04d}.{cluster}") for i, ip in enumerate(sorted_ips)
    ]

    # 构造临时表值（VALUES 语法）
    values_clause = ', '.join(["(%s, %s)"] * len(ip_hostname_pairs))
    flat_params = []
    for ip, hostname in ip_hostname_pairs:
        flat_params.extend([ip, hostname])

    sql = f"""
    UPDATE t_host AS t
    SET hostname = v.hostname
    FROM (VALUES {values_clause}) AS v(ip_addr, hostname)
    WHERE t.ip_addr = v.ip_addr
    """

    cur.execute(sql, flat_params)
    conn.commit()
    put_connection(conn)

def update_grouped_hosts(sorted_ips, prefix, cluster, group_size=24, mode="a"):
    if not sorted_ips:
        return []

    total_groups = int(len(sorted_ips) / group_size)
    usable_count = total_groups * group_size
    if usable_count == 0:
        return sorted_ips
    
    usable_ips = sorted_ips[:usable_count]
    unused_ips = sorted_ips[usable_count:]

    # 分组并构造更新映射
    all_rows = []
    cnt = 0
    for group_idx in range(total_groups):
        group_ips = usable_ips[group_idx * group_size : (group_idx + 1) * group_size]
        new_group_id = f"{prefix}{group_idx:03d}"
        if mode == "a":
            for i, ip in enumerate(group_ips):
                letter = string.ascii_lowercase[i]  # a-z
                hostname = f"{new_group_id}-{letter}.{cluster}"
                all_rows.append((ip, hostname, new_group_id))
        else:
            # 使用四位数字编号
            for i, ip in enumerate(group_ips):
                hostname = f"{prefix}-{cnt:04d}.{cluster}"
                all_rows.append((ip, hostname, new_group_id))
                cnt += 1

    # 构造批量 UPDATE SQL
    values_clause = ', '.join(['(%s, %s, %s)'] * len(all_rows))
    flat_params = []
    for ip, hostname, gid in all_rows:
        flat_params.extend([ip, hostname, gid])

    conn = get_connection()
    cur = conn.cursor()
    
    sql = f"""
    UPDATE t_host AS t
    SET hostname = v.hostname,
        group_id = v.group_id
    FROM (VALUES {values_clause}) AS v(ip_addr, hostname, group_id)
    WHERE t.ip_addr = v.ip_addr
    """

    cur.execute(sql, flat_params)
    conn.commit()
    put_connection(conn)
    return unused_ips