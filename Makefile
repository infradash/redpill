.PHONY: setup

include hack/make/*.mk

setup:
	echo "Install godep, etc."
	./hack/env.sh

test: compile
	echo "Run tests"
	${GODEP} go test ./pkg/... -v check.vv -logtostderr

compile: setup
	echo "Building redpill with LDFLAGS=$(LDFLAGS)"
	${GODEP} go build -o bin/redpill -ldflags "$(LDFLAGS)" main/redpill.go

compile-godep:
	echo "Building redpill with godep"
	${GODEP} go build -o bin/redpill -ldflags "$(LDFLAGS)" main/redpill.go

test-godep: setup
	echo "Run tests with godep"
	${GODEP} go test ./pkg/... -v check.vv -logtostderr

run-local-godep:
	PORT=5050 \
	${GODEP} go run main/redpill.go -logtostderr ${TEST_ARGS}

run-local: setup
	PORT=5050 \
	${GODEP} go run main/redpill.go -logtostderr -v=200

# Ex: make GODEP=godep ARGS=--mock=false run
run: setup
	PORT=5050 \
	${GODEP} go run main/redpill.go -logtostderr -v=200 ${TEST_ARGS}

run-80: compile-godep
	sudo bin/redpill -logtostderr -v=200 -port=80

