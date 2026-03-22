VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS := -s -w \
	-X github.com/Aliipou/cloudcostguard/cmd.Version=$(VERSION) \
	-X github.com/Aliipou/cloudcostguard/cmd.CommitSHA=$(COMMIT) \
	-X github.com/Aliipou/cloudcostguard/cmd.BuildDate=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)

.PHONY: build test lint clean docker

build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o cloudcostguard .

test:
	go test -race -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | tail -1

lint:
	golangci-lint run

clean:
	rm -f cloudcostguard coverage.out

docker:
	docker build --build-arg VERSION=$(VERSION) --build-arg COMMIT_SHA=$(COMMIT) -t cloudcostguard:$(VERSION) .

run-aws:
	go run . scan --provider aws --output table

run-azure:
	go run . scan --provider azure --output table
