FROM golang:1.20.6

LABEL maintainer="kaichao"

ARG GOPROXY=https://goproxy.io

WORKDIR /go/src
COPY go.* /go/src/
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY *.go /go/src/
# COPY misc /go/src/misc/

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    # glibc included.
    CGO_ENABLED=0 \
    go build 

FROM debian:12-slim

# install the newest version, rsync 3.2.7, openssh 9.3p1
RUN echo "deb http://deb.debian.org/debian testing main" > /etc/apt/sources.list.d/bookworm-testing.list \ 
    && apt-get update \
    && apt-get install -y openssh-client zstd \
    && apt-get clean autoclean \
    && apt-get autoremove -y \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*
COPY --from=hub.cstcloud.cn/scalebox/actuator /root/.ssh /root/.ssh

COPY --from=0 /go/src/mwa-2309.message-router /app/bin/message-router 

ENV \
    WITH_HEADERS=yes \
    ACTION_RUN=/app/bin/message-router \
    # TRACE / DEBUG / INFO / WARN / ERROR / FATAL / PANIC
    LOG_LEVEL= \
    LOCAL_MODE= \
    # 2/3/4/6/8/12/24
    NUM_NODES_PER_GROUP=24 \
    # 单次计算的指向数，可为节点的倍数
    NUM_POINTINGS_PER_CALC=24 \
    # 指向设置
    POINTING_BEGIN=1 \
    POINTING_END=1440 \
    # 每组计算dat文件数量（秒数）
    NUM_SECONDS_PER_CALC=240 \
    ALWAYS_RUNNING=yes

COPY --from=hub.cstcloud.cn/scalebox/base /usr/local/sbin /usr/local/sbin
WORKDIR /work
ENTRYPOINT ["goagent"]
