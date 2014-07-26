$(shell ./bootstrap.sh)

$(shell ./build_config.sh build_config.mk ./)

include build_config.mk

all: build  

build:
	go install -tags $(GO_BUILD_TAGS) ./...

clean:
	go clean -i ./...

test:
	go test -tags $(GO_BUILD_TAGS) ./...
