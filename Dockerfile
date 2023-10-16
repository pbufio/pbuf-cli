FROM bash:alpine3.16

WORKDIR /app

COPY ./e2e/pbuf.yaml /app
COPY ./bin/pbuf /app/pbuf

CMD ["/app/pbuf"]