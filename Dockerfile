FROM golang:alpine AS builder

RUN apk add --no-cache git
RUN mkdir -p /go/src/github.com/usermirror/config-api

WORKDIR /go/src/github.com/usermirror/config-api
COPY . .

RUN go get -d -v ./...
RUN go install -v ./...

# final stage

FROM alpine:latest

RUN apk --no-cache add ca-certificates
COPY --from=builder /go/bin/config-api /config-api

ENTRYPOINT ./config-api
EXPOSE 8888
