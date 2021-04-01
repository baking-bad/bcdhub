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

COPY cmd/metrics cmd/metrics
COPY internal internal

WORKDIR $GOPATH/src/github.com/baking-bad/bcdhub/cmd/metrics/
RUN go build -a -installsuffix cgo -o /go/bin/metrics .

# ---------------------------------------------------------------------
#  The second stage container, for running the application
# ---------------------------------------------------------------------
FROM scratch

WORKDIR /app/metrics

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/metrics /go/bin/metrics
COPY cmd/metrics/mappings /app/metrics/mappings
COPY configs/*.yml /app/metrics/

ENTRYPOINT ["/go/bin/metrics"]
