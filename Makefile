.PHONY: help build up down logs clean test

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build all Docker images
	docker-compose build

up: ## Start all services
	docker-compose up -d

down: ## Stop all services
	docker-compose down

logs: ## Show logs from all services
	docker-compose logs -f

logs-app: ## Show logs from user service
	docker-compose logs -f user-service

logs-elasticsearch: ## Show logs from Elasticsearch
	docker-compose logs -f elasticsearch

logs-logstash: ## Show logs from Logstash
	docker-compose logs -f logstash

logs-kibana: ## Show logs from Kibana
	docker-compose logs -f kibana

clean: ## Remove all containers, networks, and volumes
	docker-compose down -v --remove-orphans
	docker system prune -f

restart: ## Restart all services
	docker-compose restart

status: ## Show status of all services
	docker-compose ps

test: ## Test the application
	curl -f http://localhost:8080/health

test-elasticsearch: ## Test Elasticsearch
	curl -f http://localhost:9200/_cluster/health

test-kibana: ## Test Kibana
	curl -f http://localhost:5601/api/status

setup: ## Initial setup - build and start
	make build
	make up
	@echo "Waiting for services to be ready..."
	@sleep 30
	@echo "Testing services..."
	make test
	make test-elasticsearch
	make test-kibana
	@echo "Setup complete! Open Kibana at http://localhost:5601"
