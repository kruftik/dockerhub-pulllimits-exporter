FROM golang:1.15-alpine as builder

WORKDIR /src/app
COPY . .

RUN apk add --no-cache \
        git \
        ca-certificates \
        upx

RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
        go build -ldflags="-w -s" -mod vendor -o /app ./cmd/...

RUN upx -q /app && \
    upx -t /app

# ---

FROM scratch

WORKDIR /

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app /app

EXPOSE 8881

CMD ["/app"]
