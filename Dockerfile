FROM golang:1.17-alpine as builder

WORKDIR /app

COPY . .

RUN go build -o bin/cork

FROM alpine:3.15

ARG user=cloudbuild
ARG group=${user}
ARG uid=1000
ARG gid=1000

WORKDIR /app

COPY --from=builder /app/bin/cork .
COPY config.yaml /etc/cork/config.yaml

RUN adduser -D ${user}

USER ${user}

ENTRYPOINT ["./cork"]