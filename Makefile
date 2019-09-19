
setup:
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.17.1

lint:
	golangci-lint run

test:
	go test \
		-mod readonly \
		-race \
		-cover \
		./...

ci: test lint