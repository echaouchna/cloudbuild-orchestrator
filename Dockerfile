FROM golang:1.17-alpine as builder

WORKDIR /app

COPY . .

RUN GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o bin/cork

FROM alpine:3.15

ARG user=cloudbuild
ARG group=${user}
ARG uid=1000
ARG gid=1000

WORKDIR /app

COPY --from=builder /app/bin/cork .

RUN addgroup -g ${gid} ${group} && \
    adduser -D -u ${uid} -G ${group} ${user}

USER ${user}

ENTRYPOINT ["./cork"]