FROM alpine:latest

RUN  mkdir -p /workspace

COPY ./config.yaml /workspace

COPY ./humpback-agent /workspace/

WORKDIR /workspace

CMD ["./humpback-agent"]
