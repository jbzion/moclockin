FROM alpine

ADD release/linux/amd64/server /app/

WORKDIR /app

ENTRYPOINT ["/app/server"]