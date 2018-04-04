FROM frolvlad/alpine-glibc:alpine-3.7

MAINTAINER bobliu bobliu0909@gmail.com

RUN apk add --no-cache bash

RUN mkdir -p /opt/app/humpback-agent/conf

ADD humpback-agent /opt/app/humpback-agent/humpback-agent

ADD conf/app.conf /opt/app/humpback-agent/conf/app.conf

ADD dumb-init /dumb-init

ENTRYPOINT ["/dumb-init", "--"]

WORKDIR /opt/app/humpback-agent/

CMD ["./humpback-agent"]

EXPOSE 8500
