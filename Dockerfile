FROM golang:1.19-alpine

RUN apk add --update-cache \
    certbot

WORKDIR /app
COPY go.mod ./

RUN go mod download

COPY *.go ./

RUN go build -o /main

CMD ["/main"]