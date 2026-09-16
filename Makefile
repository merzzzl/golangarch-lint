BINARY := golangarch-lint

.PHONY: build install lint clean

build:
	go build -o $(BINARY) ./cmd/golangarch-lint

install:
	go install ./cmd/golangarch-lint

lint:
	go vet ./...
	golangci-lint run ./...
	./$(BINARY) lint .
	./$(BINARY) docs .

clean:
	rm -f $(BINARY)
