SHELL := bash
BINARY_NAME := ipfs-add
CONFIG_PKG := github.com/dimchansky/ipfs-add/config
ARTIFACTS_DIR := $(if $(ARTIFACTS_DIR),$(ARTIFACTS_DIR),bin)

PKGS ?= $(shell glide novendor)

VERSION := $(if $(TRAVIS_TAG),$(TRAVIS_TAG),$(if $(TRAVIS_BRANCH),$(TRAVIS_BRANCH),development_in_$(shell git rev-parse --abbrev-ref HEAD)))
COMMIT := $(if $(TRAVIS_COMMIT),$(TRAVIS_COMMIT),$(shell git rev-parse HEAD))
BUILD_TIME := $(shell TZ=UTC date -u '+%Y-%m-%dT%H:%M:%SZ')

M = $(shell printf "\033[32;1m▶▶▶\033[0m")
M2 = $(shell printf "\033[32;1m▶▶▶▶▶▶\033[0m")

GO_LDFLAGS := '-X "$(CONFIG_PKG).Version=$(VERSION)" -X "$(CONFIG_PKG).BuildTime=$(BUILD_TIME)" -X "$(CONFIG_PKG).GitHash=$(COMMIT)"'

.PHONY: all
all: lint test build

.PHONY: dependencies
dependencies: ; $(info $(M) retrieving dependencies…)
	@echo "$(M2) installing Glide and locked dependencies..."
	glide --version || go get -u -f github.com/Masterminds/glide
	glide install
	@echo "$(M2) installing goimports..."
	go install ./vendor/golang.org/x/tools/cmd/goimports
	@echo "$(M2) installing golint..."
	go install ./vendor/golang.org/x/lint/golint
	@echo "$(M2) installing staticcheck..."
	go install ./vendor/honnef.co/go/tools/cmd/staticcheck

.PHONY: lint
lint: ; $(info $(M) running lint tools…)
	@echo "$(M2) checking formatting..."
	@gofiles=$$(go list -f {{.Dir}} $(PKGS) | grep -v mock) && [ -z "$$gofiles" ] || unformatted=$$(for d in $$gofiles; do goimports -l $$d/*.go; done) && [ -z "$$unformatted" ] || (echo >&2 "Go files must be formatted with goimports. Following files has problem:\n$$unformatted" && false)
	@echo "$(M2) checking vet..."
	@go vet $(PKG_FILES)
	@echo "$(M2) checking staticcheck..."
	@staticcheck $(PKG_FILES)
	@echo "$(M2) checking lint..."
	@$(foreach dir,$(PKGS),golint $(dir);)

.PHONY: test
test: ; $(info $(M) running tests…)
	go test -tags=dev -timeout 40s -race -v $(PKGS)

.PHONY: cover
cover:
	mkdir -p ./${ARTIFACTS_DIR}/.cover
	go test -race -coverprofile=./${ARTIFACTS_DIR}/.cover/cover.out -covermode=atomic -coverpkg=./... $(PKGS)
	go tool cover -func=./${ARTIFACTS_DIR}/.cover/cover.out
	go tool cover -html=./${ARTIFACTS_DIR}/.cover/cover.out -o ./${ARTIFACTS_DIR}/cover.html

.PHONY: fmt
fmt: ; $(info $(M) formatting the code…)
	@echo "$(M2) formatting files..."
	@gofiles=$$(go list -f {{.Dir}} $(PKGS) | grep -v mock) && [ -z "$$gofiles" ] || for d in $$gofiles; do goimports -l -w $$d/*.go; done

.PHONY: build
build: clean ; $(info $(M) compiling…)
	go build --ldflags=$(GO_LDFLAGS) -o $(ARTIFACTS_DIR)/$(BINARY_NAME)

.PHONY: buildx
BUILD_PLATFORMS = "windows/amd64" "darwin/amd64" "linux/amd64"
buildx: clean ; $(info $(M) cross compiling…)
	for platform in $(BUILD_PLATFORMS); do \
		platform_split=($${platform//\// }); \
		GOOS=$${platform_split[0]}; \
		GOARCH=$${platform_split[1]}; \
		HUMAN_OS=$${GOOS}; \
		if [ "$$HUMAN_OS" = "darwin" ]; then \
			HUMAN_OS='macos'; \
		fi; \
		output_name=$(ARTIFACTS_DIR)/$(BINARY_NAME); \
		if [ "$$GOOS" = "windows" ]; then \
		    output_name+='.exe'; \
		fi; \
		env GOOS=$$GOOS GOARCH=$$GOARCH go build --ldflags=$(GO_LDFLAGS) -o $${output_name}; \
		if [ "$$GOOS" = "windows" ]; then \
		    pushd ${ARTIFACTS_DIR}; zip $(BINARY_NAME)-$${HUMAN_OS}-$${GOARCH}-$(VERSION).zip $(BINARY_NAME).exe; popd; \
		else \
		    pushd ${ARTIFACTS_DIR}; tar cvzf $(BINARY_NAME)-$${HUMAN_OS}-$${GOARCH}-$(VERSION).tgz $(BINARY_NAME); popd; \
		fi; \
		rm $${output_name}; \
	done

clean:
	rm -rf $(ARTIFACTS_DIR)