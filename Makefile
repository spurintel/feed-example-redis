VERSION=$(shell git describe --tags)
COMMIT=$(shell git rev-parse --short HEAD)
DATE=$(shell date +%Y-%m-%d)
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

all: format lint test bin

bin:
	echo 'Building backend version: ${VERSION}, commit: ${COMMIT}, date: ${DATE}'
	go mod download
	mkdir -p target
	go build -trimpath -a -ldflags "-X 'main.Version=${VERSION}' -X 'main.Commit=${COMMIT}' -X 'main.Date=${DATE}'" -o target/spurredis_${GOOS}_${GOARCH} ./cmd/spurredis/

bin-linux: export GOOS=linux
bin-linux: export GOARCH=amd64
bin-linux:
	echo 'Building backend version: ${VERSION}, commit: ${COMMIT}, date: ${DATE}'
	go mod download
	mkdir -p target
	go build -trimpath -a -ldflags "-X 'main.Version=${VERSION}' -X 'main.Commit=${COMMIT}' -X 'main.Date=${DATE}'" -o target/spurredis_${GOOS}_${GOARCH} ./cmd/spurredis/

test:
	go clean -testcache
	go test ./internal/...
	go test ./cmd/...

format:
	go fmt ./...

lint:
	go vet ./...

run: bin-linux
run:
	docker compose up

publish-docker-dev:
	docker buildx create --use
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm64/v8 -t spurintelligence/spurredis:dev --push --build-arg VERSION=${VERSION} --build-arg COMMIT=${COMMIT} --build-arg DATE=${DATE} .

publish-docker:
	docker buildx create --use
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm64/v8 -t spurintelligence/spurredis:latest --push --build-arg VERSION=${VERSION} --build-arg COMMIT=${COMMIT} --build-arg DATE=${DATE} .