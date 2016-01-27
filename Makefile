test:
	@godep go test -cover ./...

test-ci:
	@godep go test -cover -race -v ./...

build:
	@godep go build ./...

.PHONY: test test-ci build
