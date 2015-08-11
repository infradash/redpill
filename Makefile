.PHONY: setup

all: build

include hack/make/*.mk

setup:
	echo "Install godep, etc."
	./hack/env.sh

test: setup
	echo "Run tests"
	$(MAKE) -C pkg/aws
	$(MAKE) -C pkg/conf
	$(MAKE) -C pkg/env
	$(MAKE) -C pkg/mock
	$(MAKE) -C pkg/orchestrate
	$(MAKE) -C pkg/registry


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

