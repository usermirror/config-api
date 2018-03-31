FROM golang:1.10-alpine

RUN mkdir -p /go/src/github.com/usermirror/config-api

RUN apk update && apk add --no-cache git

ADD . /go/src/github.com/usermirror/config-api 
WORKDIR /go/src/github.com/usermirror/config-api

RUN go get .
RUN go build -o main . 

RUN apk del git mercurial

CMD ["go", "run", "main.go"]