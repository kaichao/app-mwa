# FROM presto-rocm:1.3
FROM presto-rocm-dtk-small:1.4.5

RUN apt-get update \
    && apt-get install -y zstd \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

ENV NSUB=128 \  
    RFIARGS="-time 1.0" \
    SEARCHARGS="-zmax 50 -numharm 16" \
    LOCAL_INPUT_ROOT= \
    LOCAL_OUTPUT_ROOT= \
    LD_LIBRARY_PATH=/opt/dtk/lib:$LD_LIBRARY_PATH \
    ROCR_VISIBLE_DEVICES= \
    NCPUS=1

# COPY ./code/* /app/bin/

COPY --from=hub.cstcloud.cn/scalebox/agent /usr/local /usr/local
RUN mkdir -p /work /app/bin && echo "PATH=/app/bin:/code/presto/bin:\${PATH}" >> /root/.bashrc
WORKDIR /work
ENTRYPOINT ["goagent"]
