FROM ubuntu:20.04
RUN apt-get update \
    && apt-get install -y rsync openssh-client zstd \
    && apt-get clean autoclean \
    && apt-get autoremove -y \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

COPY run.sh /app/bin/
RUN chmod +x /app/bin/run.sh

COPY --from=hub.cstcloud.cn/scalebox/agent /usr/local/bin /usr/local/bin
COPY --from=hub.cstcloud.cn/scalebox/agent /usr/local/sbin /usr/local/sbin
WORKDIR /work
ENTRYPOINT ["goagent"]
