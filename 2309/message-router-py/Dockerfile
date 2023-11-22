FROM hub.cstcloud.cn/scalebox/agent

LABEL maintainer="Jianing Tang <tangjianing@cnic.cn>"
# 安装python
RUN apt-get update \
    && apt-get install -y python3 python3-pip \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* 
COPY *.sh *.py /app/bin/
ENV WITH_HEADERS=yes
ENV ACTION_RUN=/app/bin/messageRoute.py
RUN chmod +x /app/bin/messageRoute.py

