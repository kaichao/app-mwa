FROM ubuntu:20.04
RUN apt-get update
RUN apt -y install zstd
COPY --from=hub.cstcloud.cn/scalebox/base /usr/local/sbin /usr/local/sbin
COPY run.sh /app/bin/
RUN chmod +x /app/bin/run.sh
WORKDIR /work
ENTRYPOINT ["goagent"]
