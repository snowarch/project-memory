.PHONY: build install clean test run

BINARY_NAME=pmem
INSTALL_PATH=/usr/local/bin

build:
	go build -o $(BINARY_NAME) cmd/pmem/main.go

install: build
	sudo cp $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME)
	sudo chmod +x $(INSTALL_PATH)/$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)
	go clean

test:
	go test -v ./...

run: build
	./$(BINARY_NAME)

deps:
	go mod download
	go mod tidy
