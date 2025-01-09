BINARY_NAME=adv-videoconverter
MAIN=./main.go

.PHONY: all build clean build-linux build-windows build-darwin

all: build

build:
	go mod tidy
	go build -o $(BINARY_NAME) $(MAIN)

build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-linux $(MAIN)

build-windows:
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_NAME)-windows.exe $(MAIN)

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -o $(BINARY_NAME)-darwin $(MAIN)

clean:
	rm -f $(BINARY_NAME)*
