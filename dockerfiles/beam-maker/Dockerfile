FROM mwa-vcstools-rocm-4.0.1:latest

COPY run.sh /app/bin/

ENV \
    LOCAL_INPUT_ROOT= \
    LOCAL_OUTPUT_ROOT= \
    LOCAL_CAL_ROOT= \
    KEEP_SOURCE_FILE=yes \
    KEEP_TARGET_FILE=yes \
    COMPRESS_RESULT=yes

COPY --from=hub.cstcloud.cn/scalebox/base /usr/local/sbin /usr/local/sbin
RUN mkdir -p /work/.scalebox /app/bin && echo "PATH=/app/bin:\${PATH}" >> /root/.bashrc
WORKDIR /work
ENTRYPOINT ["goagent"]
