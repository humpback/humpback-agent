FROM alpine:latest

RUN mkdir -p /workspace/config

COPY ./config/*.yaml /workspace/config

COPY ./humpback-agent /workspace/

WORKDIR /workspace

ARG VERSION

ENV VERSION=$VERSION

CMD ["./humpback-agent"]
