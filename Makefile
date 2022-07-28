.PHONY: test
test:
	@go test ./...

.PHONY: run
run:
	@go run ./cmd/geoip-processor
