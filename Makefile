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
	rm -rf vendor && ln -s _vendor/vendor vendor
	go build -o bin/ledis-server -tags '$(GO_BUILD_TAGS)' cmd/ledis-server/*
	go build -o bin/ledis-cli -tags '$(GO_BUILD_TAGS)' cmd/ledis-cli/*
	go build -o bin/ledis-benchmark -tags '$(GO_BUILD_TAGS)' cmd/ledis-benchmark/*
	go build -o bin/ledis-dump -tags '$(GO_BUILD_TAGS)' cmd/ledis-dump/*
	go build -o bin/ledis-load -tags '$(GO_BUILD_TAGS)' cmd/ledis-load/*
	go build -o bin/ledis-repair -tags '$(GO_BUILD_TAGS)' cmd/ledis-repair/*
	rm -rf vendor

test:
	rm -rf vendor && ln -s _vendor/vendor vendor
	go test --race -tags '$(GO_BUILD_TAGS)' -timeout 2m ./...
	rm -rf vendor


clean:
	go clean -i ./...

fmt:
	gofmt -w -s  . 2>&1 | grep -vE 'vendor' | awk '{print} END{if(NR>0) {exit 1}}' 

update_vendor:
	which glide >/dev/null || curl https://glide.sh/get | sh
	which glide-vc || go get -v -u github.com/sgotti/glide-vc
	rm -r vendor && mv _vendor/vendor vendor || true
	rm -rf _vendor
ifdef PKG
	glide get --strip-vendor --skip-test ${PKG}
else
	glide update --strip-vendor --skip-test
endif
	@echo "removing test files"
	glide vc --only-code --no-tests
	mkdir -p _vendor
	mv vendor _vendor/vendor