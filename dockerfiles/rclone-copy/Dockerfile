FROM ubuntu:latest

RUN apt-get update && \
    apt-get install -y curl unzip && \
    curl https://rclone.org/install.sh | bash && \
    apt-get install -y man-db && \
    apt-get clean autoclean && \
    apt-get autoremove --yes && \
    rm -rf /var/lib/{apt,dpkg,cache,log}/
COPY ./rclone /usr/bin/
RUN chown root:root /usr/bin/rclone
RUN chmod 755 /usr/bin/rclone
RUN mkdir -p /usr/local/share/man/man1
COPY ./rclone/rclone.1 /usr/local/share/man/man1/
RUN mandb
COPY rclone.conf /root/.config/rclone/rclone.conf
#COPY run.sh /app/bin/
#RUN chmod +x /app/bin/run.sh
#WORKDIR /app/bin

#CMD ["./run.sh"]
COPY --from=hub.cstcloud.cn/scalebox/base /usr/local/sbin /usr/local/sbin
COPY run.sh /app/bin/
RUN chmod +x /app/bin/run.sh

WORKDIR /work
ENTRYPOINT ["goagent"]
