FROM ubuntu:20.04
RUN apt-get update \
    && apt-get install -y rsync openssh-client zstd \
    && apt-get install -y python3 python3-pip \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*
RUN pip install psycopg2-binary
COPY --from=hub.cstcloud.cn/scalebox/base /usr/local/sbin /usr/local/sbin
COPY run.sh /app/bin/
RUN chmod +x /app/bin/run.sh
WORKDIR /work
ENTRYPOINT ["goagent"]