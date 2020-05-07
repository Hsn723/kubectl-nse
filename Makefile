WORK_DIR = ./bin
BUILD_DIR = ./artifacts
VERSION := $(shell cat VERSION)
LDFLAGS = -ldflags "-X github.com/Hsn723/kubectl-nse/cmd.CurrentVersion=${VERSION}"
OS ?= linux
ARCH ?= amd64
ifeq ($(OS), windows)
EXT = .exe
endif

clean:
	rm -rf ${WORK_DIR} ${BUILD_DIR}

setup:
	mkdir -p ${WORK_DIR} ${BUILD_DIR}

lint: clean setup
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b ${WORK_DIR} v1.26.0
	${WORK_DIR}/golangci-lint run

test: clean
	go test -race -v ./...

build: clean setup
	env GOOS=$(OS) GOARCH=$(ARCH) go build $(LDFLAGS) -o $(BUILD_DIR)/kubectl-nse-$(OS)-$(ARCH)$(EXT) .

.PHONY: clean setup lint build
