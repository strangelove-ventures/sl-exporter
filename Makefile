default: help

.PHONY: help
help: ## Print this help message
	@echo "Available make commands:"; grep -h -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: setup
setup: ## Setup the project for local development
	@cp -n config.example.yaml config.yaml || (echo "config.yaml already exists; delete it and try again" && exit 1)
	@go mod download

.PHONY: watch
watch: ## Watch for changes to build and run the server. For local development only.
	@go run github.com/cosmtrek/air

