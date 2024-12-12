TEST_PATH=./...

MOCK_STORAGE_SRC=./internal/server/storage/interface.go
MOCK_STORAGE_DST=./internal/server/storage/mocks/mocks.go

.PHONY=mock-gen
mock-gen:
	$(GOPATH)/bin/mockgen -source=$(MOCK_STORAGE_SRC) -destination=$(MOCK_STORAGE_DST)

.PHONY=build
build:
	go build -o agent cmd/agent/main.go && go build -o server cmd/server/main.go

.PHONY: test
test:
	go test -v $(TEST_PATH)


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
