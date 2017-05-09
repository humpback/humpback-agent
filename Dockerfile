FROM bashell/alpine-bash

MAINTAINER bobliu bobliu0909@newegg.com

RUN mkdir -p /opt/app/humpback-agent/conf

COPY humpback-agent /opt/app/humpback-agent/humpback-agent

COPY conf/app.conf /opt/app/humpback-agent/conf/app.conf

WORKDIR /opt/app/humpback-agent/

CMD ["./humpback-agent"]

