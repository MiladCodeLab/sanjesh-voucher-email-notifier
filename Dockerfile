# docker build -t sanjesh:latest .
FROM golang:1.26-rc-alpine AS builder

WORKDIR /app
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    git

COPY go.mod go.sum ./
RUN  go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux \
    GOARCH=amd64 go build \
    -trimpath -ldflags="-s -w" \
    -o app

FROM gcr.io/distroless/base-debian12:nonroot

USER nonroot:nonroot
WORKDIR /app

COPY --from=builder /app/app /app/app

ENTRYPOINT ["/app/app"]
