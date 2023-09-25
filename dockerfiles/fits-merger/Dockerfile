FROM mwa-vcstools-rocm-4.0.1:latest

COPY run.sh /app/bin/

ENV \
    PROJECT=G0057 \
    OUTDIR=

COPY --from=hub.cstcloud.cn/scalebox/base /usr/local/sbin /usr/local/sbin
RUN mkdir -p /work /app/bin && echo "PATH=/app/bin:\${PATH}" >> /root/.bashrc
WORKDIR /work
ENTRYPOINT ["goagent"]
