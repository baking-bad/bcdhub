# ---------------------------------------------------------------------
#  The first stage container, for building the application
# ---------------------------------------------------------------------
FROM golang:1.19-alpine as builder

ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOOS=linux

RUN apk --no-cache add ca-certificates
RUN apk add --update git

RUN mkdir -p $GOPATH/src/github.com/baking-bad/bcdhub/

COPY ./go.* $GOPATH/src/github.com/baking-bad/bcdhub/
WORKDIR $GOPATH/src/github.com/baking-bad/bcdhub
RUN go mod download

COPY cmd/indexer cmd/indexer
COPY internal internal

WORKDIR $GOPATH/src/github.com/baking-bad/bcdhub/cmd/indexer/
RUN go build -a -installsuffix cgo -o /go/bin/indexer .

# ---------------------------------------------------------------------
#  The second stage container, for running the application
# ---------------------------------------------------------------------
FROM alpine:3.16
RUN apk add curl jq

WORKDIR /app/indexer

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/indexer /go/bin/indexer
COPY configs/*.yml /app/indexer/

COPY build/indexer/entrypoint.sh build/indexer/healthcheck.sh /
RUN chmod +x /entrypoint.sh /healthcheck.sh

ENTRYPOINT ["/go/bin/indexer"]
