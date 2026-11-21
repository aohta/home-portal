.PHONY: build linux-amd64 test check integration

build:
	./scripts/build.sh

linux-amd64:
	GOOS=linux GOARCH=amd64 OUTPUT=dist/home-portal-linux-amd64 ./scripts/build.sh

test:
	go test -race ./...
	cd apps/web && npm test

check:
	@test -z "$$(gofmt -l cmd internal integration)" || (gofmt -l cmd internal integration; exit 1)
	go vet ./...
	git diff --check

integration: build
	HOME_PORTAL_TEST_BINARY="$(CURDIR)/dist/home-portal" go test -count=1 -tags integration ./integration
