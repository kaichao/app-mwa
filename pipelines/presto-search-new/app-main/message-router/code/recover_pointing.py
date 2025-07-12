#!/usr/local/bin/python3

import os
import query_db


# remove the given files on the given host
# original shell command: ssh -p ${SSH_PORT} ${DEFAULT_USER}@${from_ip} rm -rf ${LOCAL_FITS_ROOT}/mwa/24ch/${p}
def remove_pointing_files(pointing, host):
    local_fits_root = os.getenv("LOCAL_FITS_ROOT")
    ssh_port = os.getenv("SSH_PORT")
    default_user = os.getenv("DEFAULT_USER")
    cmd = f"ssh -p {ssh_port} {default_user}@{host} rm -rf {local_fits_root}/mwa/24ch/{pointing} && rm -rf /dev/shm/scalebox/mydata/mwa/dedisp/{pointing}"
    os.system(cmd)

def reset_semaphores(pointing, host, app):
    vtask_sema_name = f"host_vtask_size:local-copy:{host}"
    sys.execute("scalebox semaphore increment vtask_sema_name")
    
    vtask_sema_name = f"host_vtask_size:local-copy-unpack:{host}"
    query_db.reset_semaphore(vtask_sema_name, app)

    vtask_sema_name = pointing
    query_db.reset_semaphores(vtask_sema_name, app)

if __name__ == "__main__":
    import sys
    if len(sys.argv) != 4:
        print("Usage: python3 recover_pointings.py <pointing> <jobId> <app>")
        sys.exit(1)
    pointing = sys.argv[1]
    jobId = sys.argv[2]
    app = sys.argv[3]
    rows = query_db.get_pointing_hosts(pointing, jobId)
    # get the host with status ON
    hosts = [host for host in rows if host[3] == "ON"]
    if len(hosts) == 0:
        return

    hostname = hosts[0][2]
    ip_addr = hosts[0][1]
    remove_pointing_files(pointing, ip_addr)
    # split the hostname, remove the suffix .xxx
    hostname = hostname.split(".")[0]
    reset_semaphores(pointing, hostname, app)