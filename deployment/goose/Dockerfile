FROM golang:1.27.1-alpine AS build

ARG GOOSE_VERSION=v3.27.3

RUN go install \
    -tags="no_mysql no_sqlite3 no_ydb no_clickhouse no_mssql no_vertica no_libsql no_azuresql" \
    github.com/pressly/goose/v3/cmd/goose@${GOOSE_VERSION}

FROM alpine:3.22

RUN addgroup -S goose && adduser -S -G goose goose

COPY --from=build /go/bin/goose /usr/local/bin/goose

USER goose

ENTRYPOINT ["goose"]
