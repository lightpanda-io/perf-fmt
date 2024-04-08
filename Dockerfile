FROM golang:1.22-alpine AS builder

ADD . /src/
WORKDIR /src

ARG OS=linux
ARG ARCH=amd64

RUN CGO_ENABLED=0 GOOS=$OS GOARCH=$ARCH go build

FROM alpine:3.18

COPY --from=builder /src/perf-fmt /perf-fmt

WORKDIR /
ENTRYPOINT ["/perf-fmt"]
