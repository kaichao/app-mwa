FROM ubuntu:20.04
RUN apt-get update
RUN apt -y install zstd
# 安装python
RUN apt-get update \
    && apt-get install -y python3 python3-pip \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*
RUN pip install psycopg2-binary

ENV start=no
COPY --from=hub.cstcloud.cn/scalebox/base /usr/local/sbin /usr/local/sbin
COPY run.sh /app/bin/
RUN chmod +x /app/bin/run.sh

WORKDIR /work
ENTRYPOINT ["goagent"]