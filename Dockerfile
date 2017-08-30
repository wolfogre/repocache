FROM centos:7

COPY repocache /opt/repocache

COPY html/ /opt/html/

COPY img/ /opt/img/

COPY repo/ /opt/repo/

COPY sh/ /opt/sh/

VOLUME ["/opt/cache"]

EXPOSE 80

WORKDIR /opt/

ENTRYPOINT ["./repocache"]
