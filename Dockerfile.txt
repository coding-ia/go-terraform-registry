FROM golang:1.24.1-alpine3.21 AS builder

ARG BUILD_ID="dev"

WORKDIR /src
COPY . .

RUN go build -ldflags "-w -s -X main.version=${BUILD_ID}" -o tfrepo_server ./cmd/server
RUN go build -ldflags "-w -s" -o tfrepoctl ./cmd/cli

FROM alpine:3.21

RUN addgroup -S repouser && adduser -S repouser -G repouser
RUN mkdir /data && chown repouser:repouser /data

WORKDIR /app

COPY --from=builder /src/tfrepo_server /app
COPY --from=builder /src/tfrepoctl /app

RUN chown -R repouser:repouser /app

VOLUME /data

EXPOSE 8080

ENV BADGER_DB_PATH="/data/registry_db"

USER repouser
CMD ["/app/tfrepo_server"]
