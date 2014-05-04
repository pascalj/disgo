#!/usr/bin/make -f

SHELL=/bin/bash

all: build release

build: deps
	go build

release: clean deps golang-crosscompile
	mkdir -p build
	source golang-crosscompile/crosscompile.bash; \
	go-darwin-386 build -o build/disgo-darwin-i386; \
	go-darwin-amd64 build -o build/disgo-darwin-amd64; \
	go-linux-386 build -o build/disgo-linux-i386; \
	go-linux-amd64 build -o build/disgo-linux-amd64; \
	go-linux-arm build -o build/disgo-linux-armv6l; \
	go-freebsd-386 build -o build/disgo-freebsd-i386; \
	go-freebsd-amd64 build -o build/disgo-freebsd-amd64

golang-crosscompile:
	git clone https://github.com/davecheney/golang-crosscompile.git

deps:
	go get

clean:
	rm -rf build
	rm -f disgo
