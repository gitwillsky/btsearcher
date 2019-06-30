FROM golang:alpine AS builder

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update && \
    apk add --no-cache --virtual build-dependencies build-base gcc musl-dev git

WORKDIR /opt

COPY . .

RUN go build -ldflags '-s -w' -i cmd/btengine.go && \
    mkdir dist && \
    mv key.txt ./dist/ && \
    mv btengine ./dist/



FROM alpine:latest

MAINTAINER willsky<hdu_willsky@foxmail.com>

WORKDIR /opt

COPY --from=builder ["/opt/dist", "/opt/"]

EXPOSE 8010

ENTRYPOINT ["/opt/btengine", "-p 8010"]
