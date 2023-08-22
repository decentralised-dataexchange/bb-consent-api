recreate-cert: ## Recreates bb-consent.dev selfsigned certificate
	openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout $(CURDIR)/scripts/docker-proxy/certs/bb-consent.dev.key -out $(CURDIR)/scripts/docker-proxy/certs/bb-consent.dev.crt -subj "/C=EU/ST=Europe/L=Europe/O=bb-consent Development/OU=DevOps Department/CN=*.bb-consent.dev"

bootstrap: ## Bootstraps development environment (proxy)
	@./scripts/docker-proxy/main.sh

.DEFAULT_GOAL := help
.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
