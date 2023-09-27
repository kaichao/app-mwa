FROM mwa-vcstools-rocm-4.0.1:latest

COPY run.sh /app/bin/run.sh

ENV \
    DIR_1CH= \
    DIR_24CH= \
    ENABLE_ZSTD=

COPY --from=hub.cstcloud.cn/scalebox/base /usr/local/sbin /usr/local/sbin
RUN mkdir -p /work /app/bin && echo "PATH=/app/bin:\${PATH}" >> /root/.bashrc
WORKDIR /work
ENTRYPOINT ["goagent"]
