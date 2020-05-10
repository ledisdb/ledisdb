INSTALL_PATH ?= $(CURDIR)

$(shell ./tools/build_config.sh build_config.mk $INSTALL_PATH)

include build_config.mk

export CGO_CFLAGS
export CGO_CXXFLAGS
export CGO_LDFLAGS
export LD_LIBRARY_PATH
export DYLD_LIBRARY_PATH
export GO_BUILD_TAGS
export GO111MODULE=on

PACKAGES ?= $(shell GO111MODULE=on go list -mod=vendor ./... | grep -v /vendor/)
DIST := bin
VERSION ?= $(shell git describe --tags --always | sed 's/-/+/' | sed 's/^v//')
LDFLAGS := $(LDFLAGS) -X "main.version=$(VERSION)" -X "main.buildTags=$(GO_BUILD_TAGS)"

all: build

build: build-ledis

build-ledis:
	go build -mod=vendor -tags '$(GO_BUILD_TAGS)' -ldflags '-s -w $(LDFLAGS)' -o $(DIST)/ledis cmd/ledis/*.go

build-commands:
	go build -mod=vendor -o $(DIST)/ledis-server -tags '$(GO_BUILD_TAGS)' cmd/ledis-server/*.go
	go build -mod=vendor -o $(DIST)/ledis-cli -tags '$(GO_BUILD_TAGS)' cmd/ledis-cli/*.go
	go build -mod=vendor -o $(DIST)/ledis-benchmark -tags '$(GO_BUILD_TAGS)' cmd/ledis-benchmark/*.go
	go build -mod=vendor -o $(DIST)/ledis-dump -tags '$(GO_BUILD_TAGS)' cmd/ledis-dump/*.go
	go build -mod=vendor -o $(DIST)/ledis-load -tags '$(GO_BUILD_TAGS)' cmd/ledis-load/*.go
	go build -mod=vendor -o $(DIST)/ledis-repair -tags '$(GO_BUILD_TAGS)' cmd/ledis-repair/*.go

vet:
	go vet -mod=vendor -tags '$(GO_BUILD_TAGS)' ./...

test:
	go test -mod=vendor --race -tags '$(GO_BUILD_TAGS)' -cover -coverprofile coverage.out -timeout 10m $(PACKAGES)

clean:
	go clean -i ./...

fmt:
	gofmt -w -s  . 2>&1 | grep -vE 'vendor' | awk '{print} END{if(NR>0) {exit 1}}'

sync_vendor:
	go mod tidy -v && go mod vendor

update_vendor: sync_vendor