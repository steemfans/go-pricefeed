.PHONY: build test docker

build:
	go build -o bin/pricefeed ./cmd/pricefeed

test:
	go test ./... -count=1

docker:
	docker build -t steemfans/go-pricefeed:latest .
