FROM centos:7

COPY repocache /opt/repocache

COPY html/ /opt/

COPY repo/ /opt/

COPY sh/ /opt/

RUN chmod +x /opt/repocache

VOLUME ["/opt/cache"]

EXPOSE 80

ENTRYPOINT ["/opt/repocache"]