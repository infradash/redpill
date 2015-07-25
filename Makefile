.PHONY: setup

include hack/make/*.mk

all: build

setup:
	echo "Install godep, etc."
	./hack/env.sh

test: setup
	echo "Run tests"
	${GODEP} go test ./pkg/... -v check.vv -logtostderr

build: test
	echo "Building redpill with LDFLAGS=$(LDFLAGS)"
	${GODEP} go build -o ${BUILD_DIR}/redpill -ldflags "$(LDFLAGS)" main/redpill.go

run-local: setup
	PORT=5050 \
	${GODEP} go run main/redpill.go -logtostderr -v=200 ${TEST_ARGS}

# Ex: make GODEP=godep ARGS=--mock=false run
run: setup
	PORT=5050 \
	${GODEP} go run main/redpill.go -logtostderr -v=200 ${TEST_ARGS}

run-80: build
	sudo ${BUILD_DIR}/redpill -logtostderr -v=200 -port=80

