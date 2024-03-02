FROM mwa-vcstools-rocm-4.0.1:20240228

RUN apt-get update \
    && apt-get install -y zstd \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

ENV \
    LOCAL_INPUT_ROOT= \
    LOCAL_OUTPUT_ROOT= \
    LOCAL_CAL_ROOT= \
    KEEP_SOURCE_FILE=yes \
    LD_LIBRARY_PATH=/opt/rocm/lib:$LD_LIBRARY_PATH

COPY --from=hub.cstcloud.cn/scalebox/agent /usr/local/bin /usr/local/bin
COPY --from=hub.cstcloud.cn/scalebox/agent /usr/local/sbin /usr/local/sbin
RUN mkdir -p /work/.scalebox /app/bin && echo "PATH=/app/bin:\${PATH}" >> /root/.bashrc
WORKDIR /work
ENTRYPOINT ["goagent"]
