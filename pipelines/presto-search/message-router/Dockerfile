FROM debian:12-slim

# install the newest version, rsync 3.2.7, openssh 9.3p1
RUN echo "deb http://deb.debian.org/debian testing main" > /etc/apt/sources.list.d/bookworm-testing.list \ 
    && apt-get update \
    && apt-get install -y openssh-client zstd \
    && apt-get clean autoclean \
    && apt-get autoremove -y \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
COPY --from=hub.cstcloud.cn/scalebox/actuator /root/.ssh /root/.ssh


COPY --from=hub.cstcloud.cn/scalebox/agent /usr/local /usr/local
WORKDIR /work
ENTRYPOINT ["goagent"]