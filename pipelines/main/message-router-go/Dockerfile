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

FROM hub.cstcloud.cn/scalebox/agent

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
    NUM_SECONDS_PER_CALC=240
