BINARY  := termsnake
LDFLAGS := -s -w

.PHONY: build run test test-race vet lint fmt tidy clean

build:
	CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/termsnake

run:
	go run ./cmd/termsnake

test:
	go test ./...

test-race:
	go test -race ./...

vet:
	go vet ./...

lint:
	golangci-lint run

fmt:
	gofmt -w .

tidy:
	go mod tidy

clean:
	rm -rf bin
