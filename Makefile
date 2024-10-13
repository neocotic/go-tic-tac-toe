all: download tidy format build test

download:
	go mod download

tidy:
	go mod tidy -v

format:
	go fmt

build:
	go build -v ./...

test:
	go test -v ./...

update:
	go get -u all

vhs:
	vhs ./_examples/vhs/v0.tape
