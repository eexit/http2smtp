VERSION?=v0.1.0+dev

.PHONY: build
.DEFAULT: usage

usage:
	@echo '+-------------------------------------------------------------------------------------------+'
	@echo '| Make Usage                                                                                |'
	@echo '+-------------------------------------------------------------------------------------------+'
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "|- \033[33m%-15s\033[0m -> %s\n", $$1, $$2}'

install-sls: ## Installs serverless (via NPM)
	npm install -g serverless --ignore-scripts --loglevel error

lambda-build: ## Builds the binary
	@docker build  --build-arg version=${VERSION} --build-arg target=lambda --tag=tmp .
	@docker create --name=tmp tmp sh
	@docker cp tmp:/http2smtp ./
	@docker rm -f tmp

lambda-deploy: ## Deploys the stack
	sls deploy --verbose --config lambda-serverless.yml

lambda-remove: ## Removes the stack
	sls remove --verbose --config lambda-serverless.yml

lambda-print: ## Prints the resolved stack config
	sls print --verbose --config lambda-serverless.yml
