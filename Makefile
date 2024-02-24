.PHONY: build-apocli build-urlhandler build install-urlhandler install-apocli install

build-apocli:
	go build -o ./bin/apocli ./cmd/apocli

build-urlhandler:
	cp -R ./cmd/urlhandler/ApoCliUrlHandler.app ./bin
	go build -o ./bin/ApoCliUrlHandler.app/Contents/ApoCliUrlHandler ./cmd/urlhandler

build: build-apocli build-urlhandler

install-apocli: build-apocli
# HACK: `go install` doesn't let you override the output binary name
	go build -o $(shell go env GOPATH)/bin/apocli ./cmd/apocli

install-urlhandler: build-urlhandler
	cp -R ./bin/ApoCliUrlHandler.app /Applications/

install: install-apocli install-urlhandler