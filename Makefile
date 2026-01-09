.PHONY: all build test test-unit test-integration test-verbose test-race coverage coverage-check lint fmt fmt-check vet check ci clean deps help

BINARY_NAME := gwork
GO := go
LINTER := golangci-lint
COVERAGE_THRESHOLD := 70

all: check build

build:
	$(GO) build -o $(BINARY_NAME) .

test:
	$(GO) test -v ./...

test-unit:
	$(GO) test -v -short -tags=!integration ./...

test-integration:
	$(GO) test -v -tags=integration ./...

test-verbose:
	$(GO) test -v ./...

test-race:
	$(GO) test -race ./...

coverage:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

coverage-check:
	@$(GO) test -coverprofile=coverage.out ./... > /dev/null
	@total=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ -z "$$total" ]; then \
		echo "Error: Could not calculate coverage"; \
		exit 1; \
	fi; \
	echo "Total coverage: $$total%"; \
	if [ $$(echo "$$total < $(COVERAGE_THRESHOLD)" | bc -l) -eq 1 ]; then \
		echo "Coverage $$total% is below threshold $(COVERAGE_THRESHOLD)%"; \
		exit 1; \
	fi; \
	echo "Coverage $$total% meets threshold $(COVERAGE_THRESHOLD)%"

fmt:
	gofmt -s -w .

fmt-check:
	@test -z "$$(gofmt -s -l . | tee /dev/stderr)" || (echo "Code is not formatted. Run 'make fmt'" && exit 1)

vet:
	$(GO) vet ./...

lint:
	@which $(LINTER) > /dev/null || (echo "Installing $(LINTER)..." && go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest)
	$(LINTER) run ./...

check: fmt vet lint test

ci: fmt-check vet lint test

clean:
	$(GO) clean
	rm -f $(BINARY_NAME) coverage.out coverage.html

deps:
	$(GO) mod download
	$(GO) mod tidy

help:
	@echo "Available targets:"
	@echo "  all              - Run check and build (default)"
	@echo "  build            - Build the binary"
	@echo ""
	@echo "Testing:"
	@echo "  test             - Run all tests with verbose output"
	@echo "  test-unit        - Run only unit tests (excludes integration tests)"
	@echo "  test-integration - Run only integration tests (requires integration tag)"
	@echo "  test-verbose     - Run all tests with verbose output (same as test)"
	@echo "  test-race        - Run tests with race condition detection"
	@echo "  coverage         - Generate coverage report (HTML)"
	@echo "  coverage-check   - Check if coverage meets threshold ($(COVERAGE_THRESHOLD)%)"
	@echo ""
	@echo "Code Quality:"
	@echo "  fmt              - Format code with gofmt"
	@echo "  fmt-check        - Check formatting without modifying (for CI)"
	@echo "  vet              - Run go vet"
	@echo "  lint             - Run golangci-lint"
	@echo "  check            - Run fmt, vet, lint, and test"
	@echo "  ci               - Run fmt-check, vet, lint, and test (for CI)"
	@echo ""
	@echo "Maintenance:"
	@echo "  clean            - Remove build artifacts and coverage files"
	@echo "  deps             - Download and tidy dependencies"
	@echo "  help             - Show this help message"
