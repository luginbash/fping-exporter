VERSION = 0.1.1

build:
	gox -os="windows linux darwin" -arch="amd64 arm64" -verbose \
	    -ldflags "-X main.buildCommit=`git rev-parse --short HEAD` \
	              -X main.buildDate=`date +%Y-%m-%d` \
	              -X main.buildVersion=$(VERSION)" \
	    ./...

before_build:
	go install github.com/mitchellh/gox@latest
	
lint:
	golangci-lint run *.go
