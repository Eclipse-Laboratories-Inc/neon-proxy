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
	./bin/neon-wssubscriber

run-tests: tools mocks
	go test ./... -v
tools:
	go install github.com/golang/mock/mockgen@latest

mocks:
	mockgen -destination=pkg/solana/mocks/mock_solana_rpc_connection.go -source=pkg/solana/client.go -package=mocks solana SolanaRpcConnection
	mockgen -destination=internal/indexer/mock_solana_signs_db.go -source=internal/indexer/solana_signs_db.go -package=indexer indexer SolanaSignsDBInterface
# Empty rule for force run some targets always.
FORCE: