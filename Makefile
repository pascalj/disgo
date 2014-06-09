#!/usr/bin/make -f

OS = darwin linux freebsd windows
ARCHS = 386 amd64

all: build release

build: deps
	go build

release: clean deps
	@for arch in $(ARCHS);\
	do \
		for os in $(OS);\
		do \
			echo "Building $$os-$$arch"; \
			mkdir -p build/disgo-$$os-$$arch/; \
			GOOS=$$os GOARCH=$$arch go build -o build/disgo-$$os-$$arch/disgo; \
			cp -r public templates disgo.gcfg.sample README.md build/disgo-$$os-$$arch/; \
			tar cz -C build -f build/disgo-$$os-$$arch.tar.gz disgo-$$os-$$arch; \
		done \
	done

deps:
	go get

clean:
	rm -rf build
	rm -f disgo
