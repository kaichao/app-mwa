FROM psrfits_utils_new

RUN apt-get update \
    && apt-get install -y zstd \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY code/run.sh /app/bin/

ENV \
    LOCAL_INPUT_ROOT= \
    LOCAL_OUTPUT_ROOT= \
    DOWNSAMP_FACTOR_TIME=4 \
    KEEP_SOURCE_FILE=no

COPY --from=hub.cstcloud.cn/scalebox/agent /usr/local /usr/local
RUN mkdir -p /work/.scalebox /app/bin && echo "PATH=/app/bin:\${PATH}" >> /root/.bashrc
WORKDIR /work
ENTRYPOINT ["goagent"]
