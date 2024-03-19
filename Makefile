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

publish-docker-linux: bin-linux
publish-docker-linux:
	docker build -t spurredis:latest --platform linux/amd64 .
	docker tag spurredis:latest spurredis:$(VERSION)
	docker tag spurredis:latest spurredis:$(COMMIT)
	docker push spurredis:latest
	docker push spurredis:$(VERSION)
	docker push spurredis:$(COMMIT)