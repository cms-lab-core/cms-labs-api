FROM docker.io/library/golang:1.24.4-alpine3.22 AS builder

ARG TARGETARCH
WORKDIR /src

RUN apk add --no-cache ca-certificates git

COPY shared/go.mod shared/go.sum ./shared/
COPY clabgate/go.mod clabgate/go.sum ./clabgate/
WORKDIR /src/clabgate
RUN go mod download

WORKDIR /src
COPY shared ./shared
COPY clabgate ./clabgate
WORKDIR /src/clabgate
RUN CGO_ENABLED=0 GOOS=linux GOARCH="${TARGETARCH}" go build -ldflags="-w -s" -o /out/clabgate main.go && \
    cp start.sh /out/

FROM docker.io/library/alpine:3.24.2

WORKDIR /app
RUN apk add --no-cache ca-certificates git
COPY --from=builder /out /app/

RUN chmod +x /app/start.sh /app/clabgate

# Command to run when starting the container.
ENTRYPOINT ["/app/start.sh"]
EXPOSE 5000
