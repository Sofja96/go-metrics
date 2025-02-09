TEST_PATH=./...

MOCK_STORAGE_SRC=./internal/server/storage/interface.go
MOCK_STORAGE_DST=./internal/server/storage/mocks/mocks.go

MOCK_GRPC_SRC=./internal/proto/metrics_grpc.pb.go
MOCK_GRPC_DST=./internal/agent/export/mocks/mocks.go

.PHONY=mock-gen

mock-gen:
	$(GOPATH)/bin/mockgen -source=$(MOCK_STORAGE_SRC) -destination=$(MOCK_STORAGE_DST)
	$(GOPATH)/bin/mockgen -source=$(MOCK_GRPC_SRC) -destination=$(MOCK_GRPC_DST)

.PHONY=build
build:
	go build -o agent cmd/agent/main.go && go build -o server cmd/server/main.go

.PHONY: test
test:
	go test -v $(TEST_PATH)

.PHONY: test-without-pb
test-without-pb:
	go test -v -coverpkg=./... -coverprofile=coverage.out -covermode=count ./...
	grep -v ".pb.go" coverage.out > coverage_filtered.out
	go tool cover -func=coverage_filtered.out

.PHONY: test-with-coverage
test-with-coverage:
	go test -coverprofile=cover.out -v $(TEST_PATH)
	make --silent coverage

.PHONY: coverage
coverage:
	go tool cover -html cover.out -o cover.html
	open cover.html

.PHONY: total-coverage
total-coverage:
	go tool cover -func cover.out

.PHONY:clean
clean:
	rm -rf cover.out.html cover.out
