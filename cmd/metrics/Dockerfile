# ---------------------------------------------------------------------
#  The first stage container, for building the application
# ---------------------------------------------------------------------
FROM golang:1.13-alpine as builder

ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOOS=linux

RUN apk --no-cache add ca-certificates
RUN apk add --update git

RUN mkdir -p $GOPATH/src/github.com/aopoltorzhicky/bcdhub/

COPY ./go.* $GOPATH/src/github.com/aopoltorzhicky/bcdhub/
WORKDIR $GOPATH/src/github.com/aopoltorzhicky/bcdhub
RUN go mod download

COPY cmd/metrics cmd/metrics
COPY internal internal

WORKDIR $GOPATH/src/github.com/aopoltorzhicky/bcdhub/cmd/metrics/
RUN go build -a -installsuffix cgo -o /go/bin/metrics .

# ---------------------------------------------------------------------
#  The second stage container, for running the application
# ---------------------------------------------------------------------
FROM scratch

WORKDIR /app/metrics

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/metrics /go/bin/metrics
COPY /cmd/metrics/config.json /app/metrics/config.json

ENTRYPOINT ["/go/bin/metrics"]
