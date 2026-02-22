FROM golang:1.25.2

LABEL maintainer="kaichao"

ARG http_proxy
ARG https_proxy
ARG GOPROXY=direct

ENV HTTP_PROXY=$http_proxy
ENV HTTPS_PROXY=$https_proxy
ENV GOPROXY=$GOPROXY

# ARG GOPROXY=https://goproxy.io
RUN \
    apt update \
    && apt-get install -y git \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

WORKDIR /go/src
COPY go.* /go/src/
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY internal /go/src/internal
COPY app/router/*.go /go/src/app/router/
# COPY app/router/iopath/*.go /go/src/app/router/iopath/

WORKDIR /go/src/app/router

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    # glibc included.
    CGO_ENABLED=0 \
    go build 

FROM debian:13-slim

ENV HTTP_PROXY=
ENV HTTPS_PROXY=

# install the newest version, rsync 3.2.7, openssh 9.3p1
RUN echo "deb https://deb.debian.org/debian testing main" > /etc/apt/sources.list.d/bookworm-testing.list \ 
    && apt-get update \
    && apt-get install -y openssh-client zstd \
    && apt-get clean autoclean \
    && apt-get autoremove -y \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

# install Redis 7
RUN apt-get update \
    && apt-get install -y curl gnupg \
    && curl -fsSL https://packages.redis.io/gpg | gpg --dearmor -o /usr/share/keyrings/redis-archive-keyring.gpg \
    && echo "deb [signed-by=/usr/share/keyrings/redis-archive-keyring.gpg] https://packages.redis.io/deb bookworm main" > /etc/apt/sources.list.d/redis.list \
    && apt-get update \
    && apt-get install -y redis \
    && apt-get clean autoclean \
    && apt-get autoremove -y \
    && rm -rf /var/lib/apt/lists/* /tmp/* /var/tmp/*

RUN echo 'redis-cli() { command redis-cli -h "${REDIS_HOST:-localhost}" -p "${REDIS_PORT:-6379}" "$@"; }' >> /root/.bashrc

COPY --from=hub.cstcloud.cn/scalebox/actuator /root/.ssh /root/.ssh
COPY --from=0 /go/src/app/router/router /app/bin/

COPY dataset.yaml /

COPY app/vpath.yaml /

ENV \
    WORK_DIR= \
    ACTION_RUN=/app/bin/router \
    # TRACE / DEBUG / INFO / WARN / ERROR / FATAL / PANIC
    LOG_LEVEL= \
    NUM_OF_NODES=24 \
    DEFAULT_USER=root \
    REDIS_HOST= \
    REDIS_PORT= \
    BATCH_INSERT= 

COPY --from=hub.cstcloud.cn/scalebox/agent /usr/local /usr/local
WORKDIR /work
ENTRYPOINT ["goagent"]
