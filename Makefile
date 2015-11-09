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


build: setup
	echo "Building redpill with LDFLAGS=$(LDFLAGS)"
	${GODEP} go build -o ${BUILD_DIR}/redpill -ldflags "$(LDFLAGS)" main/redpill.go

run-local: setup
	PORT=5050 \
	${GODEP} go run main/redpill.go -logtostderr -v=200 ${TEST_ARGS}

run-local-s3: setup
	PORT=5050 \
	REDPILL_S3_REGION=zk:///code.blinker.com/aws/AWS_DEFAULT_REGION \
	REDPILL_S3_ACCESS_KEY=zk:///code.blinker.com/aws/AWS_ACCESS_KEY_ID \
	REDPILL_S3_ACCESS_TOKEN=zk:///code.blinker.com/aws/AWS_SECRET_ACCESS_KEY \
	REDPILL_S3_BUCKET=ops.blinker.com \
	${GODEP} go run main/redpill.go -logtostderr -v=200 ${TEST_ARGS}

run-qoriolabs-s3: setup
	PORT=5050 \
	REDPILL_S3_REGION=zk:///code.qoriolabs.com/aws/env/AWS_DEFAULT_REGION \
	REDPILL_S3_ACCESS_KEY=zk:///code.qoriolabs.com/aws/env/AWS_ACCESS_KEY_ID \
	REDPILL_S3_ACCESS_TOKEN=zk:///code.qoriolabs.com/aws/env/AWS_SECRET_ACCESS_KEY \
	REDPILL_S3_BUCKET=redpill.qoriolabs.com \
	${GODEP} go run main/redpill.go -logtostderr -v=200 ${TEST_ARGS}

# Ex: make GODEP=godep ARGS=--mock=false run
run: setup
	PORT=5050 \
	${GODEP} go run main/redpill.go -logtostderr -v=200 ${TEST_ARGS}

run-80: build
	sudo ${BUILD_DIR}/redpill -logtostderr -v=200 -port=80

