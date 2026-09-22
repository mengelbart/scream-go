.PHONY: all test vendor check-vendor clean

# Upstream sources vendored into the package directory, see scripts/vendor-scream.sh
VENDORED = RtpQueue.h ScreamRx.h ScreamTx.h ScreamRx.cpp ScreamTx.cpp \
           ScreamV2Tx.cpp ScreamV2TxStream.cpp LICENSE_SCREAM

all:
	go build ./...

test:
	go test -race ./...

# Refresh the vendored SCReAM sources from the submodule. Run after bumping it.
vendor:
	./scripts/vendor-scream.sh

# Fails if the vendored sources differ from what the submodule would produce.
check-vendor:
	@tmp=$$(mktemp -d); trap 'rm -rf "$$tmp"' EXIT; \
	./scripts/vendor-scream.sh "$$tmp" >/dev/null; \
	for f in $(VENDORED); do diff -u "$$f" "$$tmp/$$f" || exit 1; done; \
	echo "vendored sources are up to date"

clean:
	go clean
