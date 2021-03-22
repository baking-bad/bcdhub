# ---------------------------------------------------------------------
#  The first stage container, for building the application
# ---------------------------------------------------------------------
FROM golang:1.15-alpine as builder

ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOOS=linux

RUN apk --no-cache add ca-certificates
RUN apk add --update git

RUN mkdir -p $GOPATH/src/github.com/baking-bad/bcdhub/

COPY ./go.* $GOPATH/src/github.com/baking-bad/bcdhub/
WORKDIR $GOPATH/src/github.com/baking-bad/bcdhub
RUN go mod download

COPY cmd/graphql cmd/graphql
COPY internal internal

WORKDIR $GOPATH/src/github.com/baking-bad/bcdhub/cmd/graphql/
RUN go build -a -installsuffix cgo -o /go/bin/graphql .

# ---------------------------------------------------------------------
#  The second stage container, for running the application
# ---------------------------------------------------------------------
FROM scratch

WORKDIR /app/graphql

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/graphql /go/bin/graphql
COPY cmd/graphql/init.sql .
COPY configs/*.yml /app/graphql/

ENTRYPOINT ["/go/bin/graphql"]
