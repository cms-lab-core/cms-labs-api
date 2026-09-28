FROM docker.io/library/golang:1.24.4-alpine3.22 AS builder

ARG TARGETARCH
WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY shared/go.mod shared/go.sum ./shared/
COPY backend/go.mod backend/go.sum ./backend/
WORKDIR /src/backend
RUN go mod download

WORKDIR /src
COPY shared ./shared
COPY backend ./backend
WORKDIR /src/backend
RUN CGO_ENABLED=0 GOOS=linux GOARCH="${TARGETARCH}" go build -ldflags="-w -s" -o /out/apiserver main.go && \
    GOBIN=/out CGO_ENABLED=0 GOOS=linux GOARCH="${TARGETARCH}" \
    go install github.com/rubenv/sql-migrate/...@v1.8.1

RUN mkdir -p /out/platform && \
    cp -R platform/migrations /out/platform/ && \
    cp migrate.sh start.sh /out/

FROM docker.io/library/alpine:3.24.2

WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /out /app/

RUN chmod +x /app/migrate.sh /app/start.sh /app/apiserver

# Command to run when starting the container.
ENTRYPOINT ["/app/start.sh"]
EXPOSE 5000
