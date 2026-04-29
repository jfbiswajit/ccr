BINARY=ccr
VERSION=0.1.0
MODULE=github.com/biswajit/ccr

build:
	go build -ldflags="-X '$(MODULE)/cmd.version=$(VERSION)'" -o $(BINARY) .

build-all:
	GOOS=darwin  GOARCH=arm64  go build -ldflags="-X '$(MODULE)/cmd.version=$(VERSION)'" -o dist/$(BINARY)-mac-arm64 .
	GOOS=darwin  GOARCH=amd64  go build -ldflags="-X '$(MODULE)/cmd.version=$(VERSION)'" -o dist/$(BINARY)-mac-amd64 .
	GOOS=linux   GOARCH=amd64  go build -ldflags="-X '$(MODULE)/cmd.version=$(VERSION)'" -o dist/$(BINARY)-linux-amd64 .
	GOOS=windows GOARCH=amd64  go build -ldflags="-X '$(MODULE)/cmd.version=$(VERSION)'" -o dist/$(BINARY)-windows-amd64.exe .

test:
	go test ./... -v

clean:
	rm -f $(BINARY)
	rm -rf dist/

.PHONY: build build-all test clean
