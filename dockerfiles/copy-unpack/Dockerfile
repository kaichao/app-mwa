FROM debian:11-slim

RUN apt-get update \
    && apt-get install -y zstd \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

ENV \
    KEEP_SOURCE_FILE=yes

COPY --from=hub.cstcloud.cn/scalebox/base /usr/local/sbin /usr/local/sbin
COPY run.sh check.sh /app/bin/
WORKDIR /work

ENTRYPOINT ["goagent"]