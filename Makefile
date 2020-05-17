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
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh| sh -s -- -b ${WORK_DIR} v1.27.0
	${WORK_DIR}/golangci-lint run

test: clean
	go test -race -v $(go list ./... | grep -v "kubectl-nse/t")

setup-kind: setup
	GO111MODULE="on" go get sigs.k8s.io/kind@v0.8.1
	go install github.com/onsi/ginkgo/ginkgo
	kind create cluster --config t/kind/cluster.yaml

teardown-kind:
	kind delete cluster

kindtest: setup-kind
	ginkgo -v t

build: clean setup
	env GOOS=$(OS) GOARCH=$(ARCH) go build $(LDFLAGS) -o $(BUILD_DIR)/kubectl-nse-$(OS)-$(ARCH)$(EXT) .

.PHONY: clean setup lint test setup-kind teardown-kind kindtest build
