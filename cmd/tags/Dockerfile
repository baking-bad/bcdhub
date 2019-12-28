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

COPY cmd/tags cmd/tags
COPY internal internal

WORKDIR $GOPATH/src/github.com/aopoltorzhicky/bcdhub/cmd/tags/
RUN go build -a -installsuffix cgo -o /go/bin/tags .

# ---------------------------------------------------------------------
#  The second stage container, for running the application
# ---------------------------------------------------------------------
FROM scratch

WORKDIR /app/tags

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/tags /go/bin/tags
COPY /cmd/tags/config.json /app/tags/config.json

ENTRYPOINT ["/go/bin/tags"]
