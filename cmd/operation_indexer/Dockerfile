# ---------------------------------------------------------------------
#  The first stage container, for building the application
# ---------------------------------------------------------------------
FROM golang:1.13-alpine as builder

ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOOS=linux

RUN apk --no-cache add ca-certificates
RUN apk add --update git

RUN mkdir -p $GOPATH/src/github.com/baking-bad/bcdhub/

COPY ./go.* $GOPATH/src/github.com/baking-bad/bcdhub/
WORKDIR $GOPATH/src/github.com/baking-bad/bcdhub
RUN go mod download

COPY cmd/operation_indexer cmd/operation_indexer
COPY internal internal

WORKDIR $GOPATH/src/github.com/baking-bad/bcdhub/cmd/operation_indexer/
RUN go build -a -installsuffix cgo -o /go/bin/operation_indexer .

# ---------------------------------------------------------------------
#  The second stage container, for running the application
# ---------------------------------------------------------------------
FROM scratch

WORKDIR /app/operation_indexer

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/operation_indexer /go/bin/operation_indexer
COPY /cmd/operation_indexer/config.json /app/operation_indexer/config.json
COPY /cmd/operation_indexer/data/errors.json /app/operation_indexer/data/errors.json

ENTRYPOINT ["/go/bin/operation_indexer"]
