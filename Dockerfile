FROM bash:alpine3.16

WORKDIR /app

COPY ./e2e/pbuf.yaml /app
COPY ./bin/pbuf-cli /app/pbuf-cli

CMD ["/app/pbuf-cli"]