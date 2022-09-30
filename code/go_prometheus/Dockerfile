
FROM golang:1.17.1-buster

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build

CMD [ "/bin/sh" ]