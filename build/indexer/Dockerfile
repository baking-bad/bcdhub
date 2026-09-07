# ---------------------------------------------------------------------
#  The first stage container, for building the application
# ---------------------------------------------------------------------
FROM golang:1.26.6-alpine AS builder

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
RUN go build -tags=jsoniter_sloppy -o /go/bin/indexer .

# ---------------------------------------------------------------------
#  The second stage container, for running the application
# ---------------------------------------------------------------------
FROM alpine:3.24

WORKDIR /app/indexer

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/indexer /go/bin/indexer
COPY configs/*.yml /app/indexer/

RUN addgroup -g 10001 -S bcd && adduser -u 10001 -S -D -G bcd bcd
USER 10001:10001

ENTRYPOINT ["/go/bin/indexer"]
