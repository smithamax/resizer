
setup:
	go get -u github.com/alecthomas/gometalinter
	go get -u github.com/golang/dep/cmd/dep
	dep ensure
	gometalinter --install


lint:
	gometalinter --errors --vendor ./...

test:
	go test

ci: test lint