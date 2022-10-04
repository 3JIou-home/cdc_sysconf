
FROM golang:1.17.1-buster

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build

CMD [ "/bin/sh" ]

LABEL org.kolesa-team.image.maintainer="Vlassov Dmitriy <vlassov@vlassov.pw>" \
      org.kolesa-team.image.description="Golang MVP mysql-cdc connector for sysconf 2022."
