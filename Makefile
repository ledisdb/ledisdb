INSTALL_PATH ?= $(CURDIR)

$(shell ./bootstrap.sh)

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
	go install -tags '$(GO_BUILD_TAGS)' ./...

clean:
	go clean -i ./...

test:
	go test -tags '$(GO_BUILD_TAGS)' ./...

pytest:
	sh client/ledis-py/tests/all.sh
