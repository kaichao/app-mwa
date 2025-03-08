FROM golang:1.24.1

LABEL maintainer="kaichao"

ARG GOPROXY=https://goproxy.io

WORKDIR /go/src
COPY go.* /go/src/
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY internal /go/src/internal
COPY app/message-router/*.go /go/src/app/message-router/

WORKDIR /go/src/app/message-router

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

COPY --from=0 /go/src/app/message-router/message-router /app/bin/

COPY dataset.yaml /

ENV \
    WORK_DIR= \
    ACTION_RUN=/app/bin/message-router \
    # TRACE / DEBUG / INFO / WARN / ERROR / FATAL / PANIC
    LOG_LEVEL= \
    NUM_OF_NODES=24 \
    DEFAULT_USER=root \
    BATCH_INSERT= 

COPY --from=hub.cstcloud.cn/scalebox/agent /usr/local /usr/local
WORKDIR /work
ENTRYPOINT ["goagent"]
