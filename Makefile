build:
	go build -o ./bin/xchain

run: build
	./bin/xchain

test:
	go test -v ./...