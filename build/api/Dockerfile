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
RUN go get github.com/swaggo/swag/cmd/swag
RUN swag init --parseDependency --parseInternal --parseDepth 2
RUN go build -a -installsuffix cgo -o /go/bin/api .

WORKDIR $GOPATH/src/github.com/baking-bad/bcdhub
COPY scripts scripts

WORKDIR $GOPATH/src/github.com/baking-bad/bcdhub/scripts/
RUN cd esctl && go build -a -installsuffix cgo -o /go/bin/esctl .
RUN cd migration && go build -a -installsuffix cgo -o /go/bin/migration .
RUN cd nginx && go build -a -installsuffix cgo -o /go/bin/seo .

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

COPY --from=builder /go/bin/esctl /go/bin/esctl
COPY --from=builder /go/bin/migration /go/bin/migration
COPY --from=builder /go/bin/seo /go/bin/seo

ENTRYPOINT ["/go/bin/api"]
