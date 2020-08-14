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

COPY cmd/api cmd/api
COPY internal internal
COPY scripts scripts

WORKDIR $GOPATH/src/github.com/baking-bad/bcdhub/cmd/api/
RUN go get -u github.com/swaggo/swag/cmd/swag
RUN swag init --parseDependency
RUN go build -a -installsuffix cgo -o /go/bin/api .

WORKDIR $GOPATH/src/github.com/baking-bad/bcdhub/scripts/
RUN cd aliases && go build -a -installsuffix cgo -o /go/bin/aliases .
RUN cd es-aws && go build -a -installsuffix cgo -o /go/bin/es-aws .
RUN cd migration && go build -a -installsuffix cgo -o /go/bin/migration .
RUN cd rollback && go build -a -installsuffix cgo -o /go/bin/rollback .

# ---------------------------------------------------------------------
#  The second stage container, for running the application
# ---------------------------------------------------------------------
FROM scratch

ENV PATH="/go/bin/:$PATH"

WORKDIR /app/api

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/api /go/bin/api
COPY /cmd/api/*.yml /app/api/
COPY /cmd/api/data/errors.json /app/api/data/errors.json

COPY --from=builder /go/bin/aliases /go/bin/aliases
COPY --from=builder /go/bin/es-aws /go/bin/es-aws
COPY --from=builder /go/bin/migration /go/bin/migration
COPY --from=builder /go/bin/ml /go/bin/ml
COPY --from=builder /go/bin/rollback /go/bin/rollback

ENTRYPOINT ["/go/bin/api"]
