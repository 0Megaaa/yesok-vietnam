# ─── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.21-alpine AS builder

WORKDIR /src

# Install build dependencies.
RUN apk add --no-cache git ca-certificates

# Copy go mod files and download dependencies.
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -o /yesok-vietnam .

# ─── Final stage ─────────────────────────────────────────────────────────────
FROM alpine:3.19

LABEL maintainer="Yesok Vietnam"
EXPOSE 3601

# Install runtime dependencies.
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security.
RUN adduser -D -g '' appuser
USER appuser

WORKDIR /home/appuser

# Copy binary and assets from builder.
COPY --from=builder /yesok-vietnam .
COPY --from=builder /src/static   ./static
COPY --from=builder /src/templates ./templates
COPY --from=builder /src/html      ./html

ENTRYPOINT ["./yesok-vietnam"]
