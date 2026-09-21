.PHONY: fmt vet unit verify install

fmt:
	@test -z "$$(gofmt -l .)"

vet:
	go vet ./...

unit:
	go test ./...

verify: fmt vet unit

install:
	go install ./cmd/atlas
