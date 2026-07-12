BINARY := golangarch-lint

.PHONY: build install lint check docs clean

build:
	go build -o $(BINARY) ./cmd/golangarch-lint

install:
	go install ./cmd/golangarch-lint

lint:
	go vet ./...
	golangci-lint run ./...

check: build
	./$(BINARY) lint .

docs: build
	./$(BINARY) docs .

clean:
	rm -f $(BINARY)
