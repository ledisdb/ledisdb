INSTALL_PATH ?= $(CURDIR)

$(shell ./tools/build_config.sh build_config.mk $INSTALL_PATH)

include build_config.mk

export CGO_CFLAGS
export CGO_CXXFLAGS
export CGO_LDFLAGS
export LD_LIBRARY_PATH
export DYLD_LIBRARY_PATH
export GO_BUILD_TAGS

all: build

build:
	go build -o bin/ledis-server -tags '$(GO_BUILD_TAGS)' cmd/ledis-server/*
	go build -o bin/ledis-cli -tags '$(GO_BUILD_TAGS)' cmd/ledis-cli/*
	go build -o bin/ledis-benchmark -tags '$(GO_BUILD_TAGS)' cmd/ledis-benchmark/*
	go build -o bin/ledis-dump -tags '$(GO_BUILD_TAGS)' cmd/ledis-dump/*
	go build -o bin/ledis-load -tags '$(GO_BUILD_TAGS)' cmd/ledis-load/*
	go build -o bin/ledis-repair -tags '$(GO_BUILD_TAGS)' cmd/ledis-repair/*

test:
	go test --race -tags '$(GO_BUILD_TAGS)' -timeout 2m $$(go list ./... | grep -v -e /vendor/)


clean:
	go clean -i ./...

fmt:
	gofmt -w -s  . 2>&1 | grep -vE 'vendor' | awk '{print} END{if(NR>0) {exit 1}}'

sync_vendor:
	@which dep >/dev/null || go get -u github.com/golang/dep/cmd/dep
	dep ensure && dep prune

update_vendor:
	@which dep >/dev/null || go get -u github.com/golang/dep/cmd/dep
	dep ensure -update && dep prune