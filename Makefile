
# go tool nm ./luet | grep Commit
override LDFLAGS += -X "github.com/geaaru/time-master/cmd.BuildTime=$(shell date -u '+%Y-%m-%d %I:%M:%S %Z')"
override LDFLAGS += -X "github.com/geaaru/time-master/cmd.BuildCommit=$(shell git rev-parse HEAD)"

NAME ?= time-master
PACKAGE_NAME ?= $(NAME)
REVISION := $(shell git rev-parse --short HEAD || echo dev)
VERSION := $(shell git describe --tags || echo $(REVISION))
VERSION := $(shell echo $(VERSION) | sed -e 's/^v//g')
ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

.PHONY: all
all: deps build

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: test
test:
	GO111MODULE=on go install -mod=mod github.com/onsi/ginkgo/v2/ginkgo
	ginkgo -r -flake-attempts 3 ./...

.PHONY: coverage
coverage:
	go test ./... -race -coverprofile=coverage.txt -covermode=atomic

.PHONY: test-coverage
test-coverage:
	scripts/ginkgo.coverage.sh --codecov

.PHONY: clean
clean:
	rm -rf release/

.PHONY: deps
deps:
	go env
	# Installing dependencies...
	GO111MODULE=on go install -mod=mod golang.org/x/lint/golint
	GO111MODULE=on go install -mod=mod github.com/onsi/ginkgo/v2/ginkgo
	go get github.com/onsi/gomega/...
	ginkgo version

.PHONY: build
build:
	CGO_ENABLED=0 go build -ldflags '$(LDFLAGS)'

.PHONY: build-small
build-small:
	@$(MAKE) LDFLAGS+="-s -w" build
	upx --brute -1 $(NAME)

.PHONY: lint
lint:
	golint ./... | grep -v "be unexported"

.PHONY: vendor
vendor:
	go mod vendor

.PHONY: goreleaser-snapshot
goreleaser-snapshot:
	rm -rf dist/ || true
	goreleaser release --skip=validate,publish --snapshot --verbose
