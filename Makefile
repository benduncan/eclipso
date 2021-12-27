GO_PROJECT_NAME := eclipso-dns

# GO commands
go_build:
	@echo "\n....Building $(GO_PROJECT_NAME)"
	go build -ldflags "-s -w" -o ./bin/ ./cmd/eclipso 

go_dep_install:
	@echo "\n....Installing dependencies for $(GO_PROJECT_NAME)...."
	go get .

go_run:
	@echo "\n....Running $(GO_PROJECT_NAME)...."
	$(GOPATH)/bin/$(GO_PROJECT_NAME)

test:
	@echo "\n....Running tests for $(GO_PROJECT_NAME)...."
	ECLIPSO_LOG_IGNORE=1 go test ./pkg/backend
	ECLIPSO_LOG_IGNORE=1 go test ./pkg/config

# Project rules
build:
	$(MAKE) go_build

bench:
	ECLIPSO_LOG_IGNORE=1 go test -bench=. ./pkg/backend -count 5 -benchmem | tee benchmark.out
	benchstat benchmark.out


prof:
	ECLIPSO_LOG_IGNORE=1 go test -cpuprofile cpu.prof -memprofile mem.prof -bench=. ./pkg/backend

race:
	ECLIPSO_LOG_IGNORE=1 go test -race ./pkg/backend

run:
ifeq ($(ENV), dev)
	$(MAKE) build
	$(GOPATH)/bin/gin
else
	$(MAKE) go_build
	$(MAKE) go_run
endif

clean:
	#rm test.db
	#rm -rf ./pkg/*
	#rm -rf ./src/*
	#rm -rf ./bin/*

docker:
	@echo "\n....Building latest docker image and uploading to GCR ...."
	$(MAKE) test
	docker buildx build --push --platform linux/arm/v7,linux/arm64/v8,linux/amd64 --tag calacode/$(GO_PROJECT_NAME):latest .
	#docker tag $(GO_PROJECT_NAME) calacode/$(GO_PROJECT_NAME):latest
	#docker push calacode/$(GO_PROJECT_NAME):latest

.PHONY: docker db_seed go_build go_dep_install go_prep_install go_run build run restart historical-data
