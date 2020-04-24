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

all: build

build:
	go build -mod=vendor -o bin/ledis-server -tags '$(GO_BUILD_TAGS)' cmd/ledis-server/*
	go build -mod=vendor -o bin/ledis-cli -tags '$(GO_BUILD_TAGS)' cmd/ledis-cli/*
	go build -mod=vendor -o bin/ledis-benchmark -tags '$(GO_BUILD_TAGS)' cmd/ledis-benchmark/*
	go build -mod=vendor -o bin/ledis-dump -tags '$(GO_BUILD_TAGS)' cmd/ledis-dump/*
	go build -mod=vendor -o bin/ledis-load -tags '$(GO_BUILD_TAGS)' cmd/ledis-load/*
	go build -mod=vendor -o bin/ledis-repair -tags '$(GO_BUILD_TAGS)' cmd/ledis-repair/*

vet:
	go vet -mod=vendor -tags '$(GO_BUILD_TAGS)' ./...

test:
	go test -mod=vendor --race -tags '$(GO_BUILD_TAGS)' -cover -coverprofile coverage.out -timeout 2m $$(go list ./... | grep -v -e /vendor/)

clean:
	go clean -i ./...

fmt:
	gofmt -w -s  . 2>&1 | grep -vE 'vendor' | awk '{print} END{if(NR>0) {exit 1}}'

sync_vendor:
	go mod tidy -v && go mod vendor

update_vendor: sync_vendor