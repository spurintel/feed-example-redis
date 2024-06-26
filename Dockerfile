# Multi-stage build
FROM --platform=$BUILDPLATFORM golang:1.21 AS build
WORKDIR /src
COPY go.mod go.sum .
RUN go mod download
COPY . .
ARG TARGETOS TARGETARCH
ARG VERSION
ARG COMMIT
ARG DATE
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath  -ldflags "-X 'main.Version=$VERSION' -X 'main.Commit=$COMMIT' -X 'main.Date=$DATE'" -o /out/spurredis ./cmd/spurredis

# Final stage
FROM alpine
ENV SPUR_REDIS_API_TOKEN=$SPUR_REDIS_API_TOKEN
ENV SPUR_REDIS_CHUNK_SIZE=$SPUR_REDIS_CHUNK_SIZE
ENV SPUR_REDIS_CONCURRENT_NUM=$SPUR_REDIS_CONCURRENT_NUM
COPY --from=build /out/spurredis /root/spurredis
CMD ["/root/spurredis", "-api", "daemon"]