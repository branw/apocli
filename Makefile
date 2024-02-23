.PHONY: build-apocli build-urlhandler build install-urlhandler install-apocli install

build-apocli:
	go build -o ./bin/apocli ./bin/cli

build-urlhandler:
	go build -o ./bin/urlhandler/ApoCliUrlHandler.app/Contents/ApoCliUrlHandler ./bin/urlhandler

build: build-apocli build-urlhandler

install-apocli: build-apocli
# HACK: `go install` doesn't let you override the output binary name
	go build -o $(shell go env GOPATH)/bin/apocli ./bin/cli

install-urlhandler: build-urlhandler
	cp -R ./bin/urlhandler/ApoCliUrlHandler.app /Applications/

install: install-apocli install-urlhandler