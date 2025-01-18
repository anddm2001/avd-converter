BINARY_NAME=adv-videoconverter
MAIN=./main.go

.PHONY: all build clean build-macos

all: build

build:
	go mod tidy
	go build -o $(BINARY_NAME) $(MAIN)

build-macos:
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-darwin $(MAIN)

clean:
	rm -f $(BINARY_NAME)*
