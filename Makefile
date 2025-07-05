build:
	go build -o ./bin/xchain

run: build
	./bin/xchain

test:
	go test -v $(if $(FILE),$(FILE),./...) $(if $(FUNC),-run $(FUNC))