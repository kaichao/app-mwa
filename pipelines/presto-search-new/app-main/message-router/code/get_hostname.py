#!/usr/local/bin/python3
import os
import sys
from query_db import get_host_by_ip

if __name__ == "__main__":
    # print(sys.argv[1])
    hostname = get_host_by_ip(sys.argv[1])
    print(hostname[0][0])
