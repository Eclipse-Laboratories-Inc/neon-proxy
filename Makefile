VERSION_VAR := github.com/neonlabsorg/neon-proxy/pkg/service.Version
VERSION_BRANCH := $($(shell git branch | grep \* | cut -d ' ' -f2 | sed 's/(//'))
VERSION := $(shell git describe --long --abbrev=8 $(VERSION_BRANCH))
GOBUILD_VERSION_ARGS := "-X $(VERSION_VAR)=v$(VERSION).$(VERSION_BRANCH)"
BINARIES = $(sort $(addprefix bin/, $(notdir $(wildcard cmd/*))))

.PHONY: build run

build: $(BINARIES)

bin/%: FORCE
	$(info Build $(@F))
	go build -o $@ -ldflags $(GOBUILD_VERSION_ARGS) "./cmd/$(@F)"

run: build
	./bin/neon-proxy

run-proxy:
	./bin/neon-proxy

run-indexer:
	./bin/neon-indexer

run-mempool:
	./bin/neon-mempool

run-subscriber:
	./bin/neon-subscriber

run-wssubscriber:
	sh run-neon-wssubscriber.sh

# Empty rule for force run some targets always.
FORCE:
