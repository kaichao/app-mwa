FROM mwa-vcstools-rocm-4.0.1:latest

RUN apt-get update \
    && apt-get install -y zstd \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY run.sh /app/bin/run.sh

ENV \
    LOCAL_INPUT_ROOT= \
    LOCAL_OUTPUT_ROOT= \
    KEEP_SOURCE_FILE=yes \
    COMPRESS_RESULT=yes \
    ENABLE_ZSTD=

COPY --from=hub.cstcloud.cn/scalebox/base /usr/local/sbin /usr/local/sbin
RUN mkdir -p /work /app/bin && echo "PATH=/app/bin:\${PATH}" >> /root/.bashrc
WORKDIR /work
ENTRYPOINT ["goagent"]
