URL=https://api.github.com/repos/eexit/http2smtp/releases/latest

.DEFAULT: usage

usage:
	@echo '+-------------------------------------------------------------------+'
	@echo '| Make Usage                                                        |'
	@echo '+-------------------------------------------------------------------+'
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "|- \033[33m%-15s\033[0m -> %s\n", $$1, $$2}'

install-sls: ## Installs serverless (via NPM)
	npm install -g serverless --ignore-scripts --loglevel error

config: ## Prints the resolved stack config
	sls print --verbose

download: ## Gets the latest version of the binary
	curl -sL $(URL) | jq -r '.assets[].browser_download_url' | grep lambda | grep -v md5 | xargs wget -qO- | tar xvz http2smtp-lambda

deploy: ## Deploys the stack
	sls deploy --verbose

remove: ## Removes the stack
	sls remove --verbose
