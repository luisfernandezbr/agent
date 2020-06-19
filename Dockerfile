FROM golang:alpine as builder

RUN apk add git
RUN mkdir -p $GOPATH/src/github.com/pinpt/agent.next
WORKDIR $GOPATH/src/github.com/pinpt/agent.next
COPY . .
ENV GIT_TERMINAL_PROMPT 1
RUN go build -ldflags "-s -w" -o /bin/agent.next main.go

FROM alpine:edge
RUN apk add openssl-dev cyrus-sasl-dev ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /
COPY --from=builder /bin/agent.next /bin/agent.next
