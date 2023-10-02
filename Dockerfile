ARG GOLANG_VERSION="1.21.1"

FROM golang:${GOLANG_VERSION}-alpine as builder
ARG LDFLAGS
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /go/src/github.com/z0rr0/tgtpgybot
COPY . .
RUN echo "LDFLAGS = $LDFLAGS"
RUN GOOS=linux go build -ldflags "$LDFLAGS" -o ./tgtpgybot

FROM alpine:3.18
LABEL org.opencontainers.image.authors="me@axv.email" \
        org.opencontainers.image.url="https://hub.docker.com/repository/docker/z0rr0/tgtpgybot" \
        org.opencontainers.image.documentation="github.com/z0rr0/tgtpgybot" \
        org.opencontainers.image.source="github.com/z0rr0/tgtpgybot" \
        org.opencontainers.image.licenses="MIT" \
        org.opencontainers.image.title="TgTGPYBot" \
        org.opencontainers.image.description="Telegram Yandex GPT bot"
COPY --from=builder /go/src/github.com/z0rr0/tgtpgybot/tgtpgybot /bin/
RUN chmod 0755 /bin/tgtpgybot

VOLUME ["/data/tgtpgybot/"]
ENTRYPOINT ["/bin/tgtpgybot"]
CMD ["-config", "/data/tgtpgybot/config.json"]
