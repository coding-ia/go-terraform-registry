FROM golang:1.24.1-alpine3.21 AS builder

ARG BUILD_ID="dev"

WORKDIR /src
COPY . .

RUN go build -ldflags "-w -s -X main.version=${BUILD_ID}" -o tfrepo_server ./cmd/server
RUN go build -ldflags "-w -s" -o tfrepoctl ./cmd/cli

FROM alpine:3.21

RUN addgroup -S repouser && adduser -S repouser -G repouser
RUN mkdir /data && chown repouser:repouser /data
RUN mkdir /store && chown repouser:repouser /store

WORKDIR /app

COPY --from=builder /src/tfrepo_server /app
COPY --from=builder /src/tfrepoctl /app
COPY --from=builder /src/migrations /app/migrations

RUN chown -R repouser:repouser /app

VOLUME /data
VOLUME /store

EXPOSE 8080

ENV BADGER_DB_PATH="/data/registry_db"
ENV LOCAL_STORAGE_ASSETS_PATH="/store"

USER repouser
CMD ["/app/tfrepo_server"]
