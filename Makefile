.PHONY: build
# build
build:
	mkdir -p bin/ && go build -o ./bin/pbuf .

.PHONY: build-in-docker
# build in docker
build-in-docker:
	docker run --rm \
      -v ".:/app" \
      -v "./bin:/app/bin" \
      -v "${HOME}/.netrc:/root/.netrc" \
      -w /app \
      golang:1.21.1 \
      sh -c "CGO_ENABLED=0 GOOS=linux make build"

.PHONY: docker
# docker
docker:
	docker build -t pbuf:latest .

.PHONY: lint
# lint
lint:
	golangci-lint run -v --timeout 10m

# show help
help:
	@echo ''
	@echo 'Usage:'
	@echo ' make [target]'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
	helpMessage = match(lastLine, /^# (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")-1); \
			helpMessage = substr(lastLine, RSTART + 2, RLENGTH); \
			printf "\033[36m%-22s\033[0m %s\n", helpCommand,helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help
