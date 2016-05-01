INSTALL_PATH ?= $(CURDIR)

$(shell ./tools/build_config.sh build_config.mk $INSTALL_PATH)

include build_config.mk

export CGO_CFLAGS
export CGO_CXXFLAGS
export CGO_LDFLAGS
export LD_LIBRARY_PATH
export DYLD_LIBRARY_PATH
export GO_BUILD_TAGS

GO=GO15VENDOREXPERIMENT="1" go

all: build  

build:
	$(GO) build -o bin/ledis-server -tags '$(GO_BUILD_TAGS)' cmd/ledis-server/*
	$(GO) build -o bin/ledis-cli -tags '$(GO_BUILD_TAGS)' cmd/ledis-cli/*

build_all: build
	$(GO) build -o bin/ledis-benchmark -tags '$(GO_BUILD_TAGS)' cmd/ledis-benchmark/*
	$(GO) build -o bin/ledis-dump -tags '$(GO_BUILD_TAGS)' cmd/ledis-dump/*
	$(GO) build -o bin/ledis-load -tags '$(GO_BUILD_TAGS)' cmd/ledis-load/*
	$(GO) build -o bin/ledis-repair -tags '$(GO_BUILD_TAGS)' cmd/ledis-repair/*
	
test:
	# use vendor for test
	rm -rf vendor && ln -s cmd/vendor vendor
	$(GO) test --race -tags '$(GO_BUILD_TAGS)' ./... 2>&1 | grep -vE 'vendor'
	rm -rf vendor


clean:
	$(GO) clean -i ./...

fmt:
	gofmt -w -s  . 2>&1 | grep -vE 'vendor' | awk '{print} END{if(NR>0) {exit 1}}' 

deps:
	# see https://github.com/coreos/etcd/blob/master/scripts/updatedep.sh
	rm -rf Godeps vendor cmd/vendor
	mkdir -p cmd/vendor
	ln -s cmd/vendor vendor
	godep save ./...
	rm -rf cmd/Godeps
	rm vendor
	mv Godeps cmd/