VERBOSE?=0
PARALLEL?=4
VENDOR?=0


ifeq (${VENDOR}, 1)
VENDOR_DEP=vendor
endif


build: ${VENDOR_DEP} bin
ifeq (${VENDOR}, 1)
	go build -mod vendor -o bin/appmeta ./cmd/server
else
	go build -o bin/appmeta ./cmd/server
endif

bin:
	@mkdir -p bin

vendor:
	go mod vendor

checked_build:
	go vet ./cmd/server
	go build -race -mod vendor -o bin/appmeta ./cmd/server

test: ${VENDOR_DEP}
ifeq (${VENDOR}, 1)
	go test -v -race -parallel ${PARALLEL} -mod vendor ./pkg/metadata/...
else
	go test -v -race -parallel ${PARALLEL} ./pkg/metadata/...
endif

doc:
	@mkdir -p docs
	godoc  pkg/metadata... > ./docs/doc.html
	godoc  pkg/middleware >> ./docs/doc.html


clean:
	@rm -rf bin
	@rm -rf vendor


.phony: build bin
