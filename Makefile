SHELL := bash
.ONESHELL:

VER=$(shell git describe --tags --always --dirty)
GO=$(shell which go)
GOMOD=$(GO) mod
GOFMT=$(GO) fmt
GOBUILD=$(GO) build -trimpath -mod=readonly -ldflags "-X main.version=$(VER) -s -w -buildid="

dir:
	@if [ ! -d bin ]; then mkdir -p bin; fi

mod:
	@$(GOMOD) download

format:
	@$(GOFMT) ./...

build/linux/mips64: dir mod
	export CGO_ENABLED=0
	export GOOS=linux
	export GOARCH=mips64
	$(GOBUILD) -o bin/uptime-linux-$(VER:v%=%)-mips64 main.go

build/linux/armv7: dir mod
	export CGO_ENABLED=0
	export GOOS=linux
	export GOARCH=arm
	export GOARM=7
	$(GOBUILD) -o bin/uptime-linux-$(VER:v%=%)-armv7 main.go

build/linux/arm64: dir mod
	export CGO_ENABLED=0
	export GOOS=linux
	export GOARCH=arm64
	$(GOBUILD) -o bin/uptime-linux-$(VER:v%=%)-arm64 main.go

build/linux/386: dir mod
	export CGO_ENABLED=0
	export GOOS=linux
	export GOARCH=386
	$(GOBUILD) -o bin/uptime-linux-$(VER:v%=%)-386 main.go

build/linux/amd64: dir mod
	export CGO_ENABLED=0
	export GOOS=linux
	export GOARCH=amd64
	$(GOBUILD) -o bin/uptime-linux-$(VER:v%=%)-amd64 main.go

build/linux: build/linux/mips64 build/linux/armv7 build/linux/arm64 build/linux/386 build/linux/amd64

build/darwin/arm64: dir mod
	export CGO_ENABLED=0
	export GOOS=darwin
	export GOARCH=arm64
	$(GOBUILD) -o bin/uptime-darwin-$(VER:v%=%)-arm64 main.go

build/darwin/amd64: dir mod
	export CGO_ENABLED=0
	export GOOS=darwin
	export GOARCH=amd64
	$(GOBUILD) -o bin/uptime-darwin-$(VER:v%=%)-amd64 main.go

build/darwin: build/darwin/arm64 build/darwin/amd64

build/windows/amd64: dir mod
	export CGO_ENABLED=0
	export GOOS=windows
	export GOARCH=amd64
	$(GOBUILD) -o bin/uptime-windows-$(VER:v%=%)-amd64 main.go

build/windows: build/windows/amd64

build: build/linux build/darwin build/windows

clean:
	@rm -rf bin

all: format build
