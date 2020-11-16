# ---------------------------------------------------------------------
#  The first stage container, for building golang application
# ---------------------------------------------------------------------
FROM golang:1.13-alpine as app_builder

ENV CGO_ENABLED=0
ENV GO111MODULE=on
ENV GOOS=linux

RUN apk --no-cache add ca-certificates
RUN apk add --update git

RUN mkdir -p $GOPATH/src/github.com/baking-bad/bcdhub/

COPY ./go.* $GOPATH/src/github.com/baking-bad/bcdhub/
WORKDIR $GOPATH/src/github.com/baking-bad/bcdhub
RUN go mod download

COPY cmd/compiler cmd/compiler
COPY internal internal

WORKDIR $GOPATH/src/github.com/baking-bad/bcdhub/cmd/compiler/
RUN go build -a -installsuffix cgo -o /go/bin/compiler .

# ---------------------------------------------------------------------
#  The second stage container, for running the application
# ---------------------------------------------------------------------
FROM node:12.18.3-buster

RUN apt-get update
RUN apt-get install -y wget curl

# ligo
RUN wget https://ligolang.org/bin/linux/ligo && chmod +x ./ligo
RUN cp ./ligo /usr/local/bin

# smartPy
RUN wget https://smartpy.io/cli/install.sh
RUN echo y | sh install.sh

# tezos-client
RUN wget https://github.com/baking-bad/tezos-client/releases/download/24.8.20/tezos-client && chmod +x ./tezos-client
RUN cp ./tezos-client /usr/local/bin/

# golang application
ENV PATH="/go/bin/:$PATH"

WORKDIR /app/compiler

COPY --from=app_builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=app_builder /go/bin/compiler /go/bin/compiler
COPY configs/*.yml /app/compiler/

ENTRYPOINT ["/go/bin/compiler"]
