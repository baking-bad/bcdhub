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

COPY cmd/api cmd/api
COPY internal internal

WORKDIR $GOPATH/src/github.com/baking-bad/bcdhub/cmd/api/
RUN go build -a -installsuffix cgo -o /go/bin/api .

# ---------------------------------------------------------------------
#  The second stage container, for running the application
# ---------------------------------------------------------------------
FROM scratch

ENV PATH="/go/bin/:$PATH"

WORKDIR /app/api

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/api /go/bin/api
COPY configs/*.yml /app/api/
COPY cmd/api/data/tzip-16-schema.json /app/api/data/tzip-16-schema.json

ENTRYPOINT ["/go/bin/api"]
