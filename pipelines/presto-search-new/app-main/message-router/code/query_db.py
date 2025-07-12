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
    cur.execute("SELECT id, hostname FROM t_host WHERE group_id = %s AND status = 'ON'", (group_id,))
    rows = cur.fetchall()
    put_connection(conn)
    return rows

def get_hosts_likely(group_id):
    conn = get_connection()
    cur = conn.cursor()
    cur.execute("SELECT id, hostname, ip_addr, parameters FROM t_host WHERE group_id ~ %s AND status = 'ON'", (group_id,))
    rows = cur.fetchall()
    put_connection(conn)
    return rows

def get_hosts_likely_ordered(copy_id, unpack_id, dedisp_id, group_id):
    sql = """
    SELECT 
    h.id AS id,
    h.hostname AS host_name,
    COALESCE(SUM(CASE WHEN t.job = %s THEN 1 ELSE 0 END), 0) AS copy_alloc,
    COALESCE(SUM(CASE WHEN t.job = %s THEN 1 ELSE 0 END), 0) AS unpack_alloc,
    COALESCE(SUM(CASE WHEN t.job = %s THEN 1 ELSE 0 END), 0) AS dedisp_alloc
    FROM 
        t_host h
    LEFT JOIN 
        t_task t ON h.id = t.to_host
    AND 
        t.status_code in (-1, -2, -3)
    WHERE 
        h.group_id ~ %s
        AND h.status = 'ON'
    GROUP BY 
        h.hostname, h.id
    ORDER BY 
        copy_alloc, unpack_alloc, dedisp_alloc, host_name;
    """
    conn = get_connection()
    cur = conn.cursor()
    cur.execute(sql, (copy_id, unpack_id, dedisp_id, group_id))
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

def get_host_by_ip(ip):
    conn = get_connection()
    cur = conn.cursor()
    cur.execute("SELECT id, hostname FROM t_host WHERE ip_addr = %s AND status = 'ON'", (ip,))
    rows = cur.fetchall()
    put_connection(conn)
    return rows

def get_host_slots(host):
    # 从t_slot表中获取当前host每个job的slot数量
    conn = get_connection()
    cur = conn.cursor()
    cur.execute("SELECT job, COUNT(*) FROM t_slot WHERE host = %s GROUP BY job ORDER BY job", (host,))
    rows = cur.fetchall()
    put_connection(conn)
    # 将结果转换为字典
    host_slots = {}
    for row in rows:
        host_slots[row[0]] = row[1]
    return host_slots

# 从t_task中查询job为job_id,status_code为status_code，to_host为host的记录数量
# to_host可能为None,表示不限制
def get_task_by_job(job_id, status_code, host=None):
    conn = get_connection()
    cur = conn.cursor()
    if host is None:
        cur.execute("SELECT COUNT(*) FROM t_task WHERE job = %s AND status_code = %s", (job_id, status_code))
    else:
        cur.execute("SELECT COUNT(*) FROM t_task WHERE job = %s AND status_code = %s AND to_host = %s", (job_id, status_code, host))
    rows = cur.fetchall()
    put_connection(conn)
    return rows

def get_pointing_hosts(pointing, jobId):
    conn = get_connection()
    cur = conn.cursor()
    # get the host of the job local-copy of the given app with the body pointing
    sql = """
        SELECT t_host.id, t_host.ip_addr, t_host.hostname, t_host.status
        FROM t_host, t_task
        WHERE t_task.job = %s
        AND t_task.body = %s
        AND t_host.id = t_task.to_host
        """
    cur.execute(sql, (jobId, pointing))

    rows = cur.fetchall()
    put_connection(conn)
    # return only 1 row
    return rows


def reset_semaphores(name, app):
    sql = """
        UPDATE t_semaphore
        SET value = value0
        WHERE name ~ %s
        AND app = %s
        """
    conn = get_connection()
    cur = conn.cursor()
    cur.execute(sql, (name, app))
    conn.commit()
    put_connection(conn)
    
def reset_semaphore(name, app):
    sql = """
        UPDATE t_semaphore
        SET value = value0
        WHERE name = %s
        AND app = %s
        """
    conn = get_connection()
    cur = conn.cursor()
    cur.execute(sql, (name, app))
    conn.commit()
    put_connection(conn)
