.PHONY: _pwd_prompt dec enc
DEFAULT: compile-godep test-godep


################################# Utilities for Encryption / Decryption ###########################################

# 'private' task for echoing instructions
_pwd_prompt: mk_dirs

# Make directories based the file paths
mk_dirs:
	@mkdir -p encrypt decrypt ;

# Decrypt files in the encrypt/ directory
decrypt: _pwd_prompt
	@echo "Decrypt the files in a given directory (those with .cast5 extension)."
	@read -p "Source directory: " src && read -p "Password: " password ; \
	mkdir -p decrypt/$${src} && echo "\n" ; \
	for i in `ls encrypt/$${src}/*.cast5` ; do \
		echo "Decrypting $${i}" ; \
		openssl cast5-cbc -d -in $${i} -out decrypt/$${src}/`basename $${i%.*}` -pass pass:$${password}; \
		chmod 600 decrypt/$${src}/`basename $${i%.*}` ; \
	done ; \
	echo "Decrypted files are in decrypt/$${src}"

# Encrypt files in the decrypt/ directory
encrypt: _pwd_prompt
	@echo "Encrypt the files in a directory using a password you specify.  A directory will be created under /encrypt."
	@read -p "Source directory name: " src && read -p "Password: " password && echo "\n"; \
	mkdir -p encrypt/`basename $${src}` ; \
	echo "Encrypting $${src} ==> encrypt/`basename $${src}`" ; \
	for i in `ls $${src}` ; do \
		echo "Encrypting $${src}/$${i}" ; \
		openssl cast5-cbc -e -in $${src}/$${i} -out encrypt/`basename $${src}`/$${i}.cast5 -pass pass:$${password}; \
	done ; \
	echo "Encrypted files are in encrypt/`basename $${src}`"



##################################################### Builds  #######################################################

GIT_REPO:=`git config --get remote.origin.url | sed -e 's/[\/&]/\\&/g'`
GIT_TAG:=`git describe --abbrev=0 --tags`
GIT_BRANCH=`git rev-parse --abbrev-ref HEAD`
GIT_COMMIT_HASH:=`git rev-list --max-count=1 --reverse HEAD`
GIT_COMMIT_MESSAGE:=`git log -1 --format="%h,%an,%s"`
BUILD_TIMESTAMP:=`date +"%Y-%m-%d-%H:%M"`


LDFLAGS:=\
-X github.com/qorio/omni/version.gitRepo $(GIT_REPO) \
-X github.com/qorio/omni/version.gitTag $(GIT_TAG) \
-X github.com/qorio/omni/version.gitBranch $(GIT_BRANCH) \
-X github.com/qorio/omni/version.gitCommitHash $(GIT_COMMIT_HASH) \
-X github.com/qorio/omni/version.gitCommitMessage \"$(GIT_COMMIT_MESSAGE)\" \
-X github.com/qorio/omni/version.buildTimestamp $(BUILD_TIMESTAMP) \
-X github.com/qorio/omni/version.buildNumber $(BUILD_NUMBER) \


setup:
	echo "Install godep, etc."
	./bin/env.sh

test: compile
	echo "Run tests"
	go test ./pkg/... -v check.vv -logtostderr

compile: setup
	echo "Building redpill with LDFLAGS=$(LDFLAGS)"
	go build -o bin/redpill -ldflags "$(LDFLAGS)" main/redpill.go

compile-godep:
	echo "Building redpill with godep"
	godep go build -o bin/redpill -ldflags "$(LDFLAGS)" main/redpill.go

test-godep: setup
	echo "Run tests with godep"
	godep go test ./pkg/... -v check.vv -logtostderr

run-local-godep:
	PORT=5050 \
	godep go run main/redpill.go -logtostderr

run-local: setup
	PORT=5050 \
	go run main/redpill.go -logtostderr -v=200

run-80: compile-godep
	sudo bin/redpill -logtostderr -v=200 -port=80

