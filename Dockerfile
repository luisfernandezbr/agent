FROM golang:alpine as builder

RUN apk add git
RUN mkdir -p $GOPATH/src/github.com/pinpt/agent
WORKDIR $GOPATH/src/github.com/pinpt/agent
COPY . .
ENV GIT_TERMINAL_PROMPT 1
RUN go build -ldflags "-s -w -X 'main.date=`date -R`' -X main.version=`git tag --sort=-v:refname | head -n 1` -X main.commit=`git rev-parse HEAD`" -o /bin/agent main.go

FROM alpine:edge
RUN apk add openssl-dev cyrus-sasl-dev ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /
COPY --from=builder /bin/agent /bin/agent

ENTRYPOINT [ "/bin/agent" ]
