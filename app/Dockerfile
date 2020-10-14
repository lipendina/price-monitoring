FROM golang:1.15

ENV GO111MODULE=on

ADD . /app
WORKDIR /app

RUN go mod download

WORKDIR /app/cmd
RUN go build

WORKDIR /app
CMD ["./cmd/cmd"]