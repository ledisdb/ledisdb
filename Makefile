GO_BUILD_FLAG += leveldb

all: build  

build:
	go install -tags $(GO_BUILD_FLAG) ./...

clean:
	go clean -i ./...

test:
	go test -tags $(GO_BUILD_FLAG) ./...
