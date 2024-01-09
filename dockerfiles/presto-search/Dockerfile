from presto-rocm:1.3

ENV NSUB=128 \  
    RFIARGS= \
    SEARCHARGS= \
    SETTINGS_FILE= 

COPY ./scripts/* /app/bin/

COPY --from=hub.cstcloud.cn/scalebox/base /usr/local/sbin /usr/local/sbin
RUN mkdir -p /work /app/bin && echo "PATH=/app/bin:/code/presto/bin:\${PATH}" >> /root/.bashrc
WORKDIR /work
ENTRYPOINT ["goagent"]


