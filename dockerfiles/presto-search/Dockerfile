FROM presto-rocm:1.3

ENV NSUB=128 \  
    RFIARGS="-time 1.0" \
    SEARCHARGS='"-zmax 50 -numharm 16"' \
    LOCAL_INPUT_ROOT= \
    LOCAL_OUTPUT_ROOT= \
    ROCR_VISIBLE_DEVICES= 

COPY ./scripts/* /app/bin/

COPY --from=hub.cstcloud.cn/scalebox/base /usr/local/sbin /usr/local/sbin
RUN mkdir -p /work /app/bin && echo "PATH=/app/bin:/code/presto/bin:\${PATH}" >> /root/.bashrc
WORKDIR /work
ENTRYPOINT ["goagent"]


