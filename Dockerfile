# Multi-stage build
FROM --platform=$BUILDPLATFORM golang:1.21 AS build
WORKDIR /src
COPY go.mod go.sum .
RUN go mod download
COPY . .
ARG TARGETOS
ARG TARGETARCH
ARG VERSION
ARG COMMIT
ARG DATE
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags "-X 'main.Version=$VERSION' -X 'main.Commit=$COMMIT' -X 'main.Date=$DATE'" -o /out/spurredis ./cmd/spurredis

# Final stage
FROM alpine
# Define ARG with default values for build-time configuration
ARG SPUR_REDIS_CHUNK_SIZE=5000
ARG SPUR_REDIS_CONCURRENT_NUM=10

# Set non-sensitive environment variables with defaults
ENV SPUR_REDIS_CHUNK_SIZE=$SPUR_REDIS_CHUNK_SIZE
ENV SPUR_REDIS_CONCURRENT_NUM=$SPUR_REDIS_CONCURRENT_NUM

# Note: SPUR_REDIS_API_TOKEN is intentionally not set here as it's sensitive
# and should be provided at runtime via docker-compose environment variables

COPY --from=build /out/spurredis /root/spurredis
CMD ["/root/spurredis", "-api", "daemon"]