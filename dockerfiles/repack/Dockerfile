FROM ubuntu:20.04
RUN apt-get update \
    && apt-get install -y rsync openssh-client zstd \
    && apt-get clean autoclean \
    && apt-get autoremove -y \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY --from=hub.cstcloud.cn/scalebox/base /usr/local/sbin /usr/local/sbin
COPY run.sh /app/bin/
RUN chmod +x /app/bin/run.sh
WORKDIR /work
ENTRYPOINT ["goagent"]
