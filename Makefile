.PHONY: help install deploy deploy-frontend deploy-backend clean status logs restart

# Default backend to deploy
BACKEND ?= go

# Required environment variables
ifndef DOMAIN
$(error DOMAIN is not set. Usage: make deploy DOMAIN=example.com EMAIL=admin@example.com BACKEND=go)
endif

ifndef EMAIL
$(error EMAIL is not set. Usage: make deploy DOMAIN=example.com EMAIL=admin@example.com BACKEND=go)
endif

# Deployment paths
DEPLOY_DIR := /opt/pebble
FRONTEND_DIR := $(DEPLOY_DIR)/frontend
BACKEND_DIR := $(DEPLOY_DIR)/backend
WWW_ROOT := /var/www/pebble
SCRIPT_DIR := $(shell pwd)/scripts

export DOMAIN
export EMAIL
export BACKEND

help:
	@echo "Pebble Deployment System"
	@echo ""
	@echo "Usage:"
	@echo "  make deploy DOMAIN=example.com EMAIL=admin@example.com BACKEND=go|py|kt|rs"
	@echo ""
	@echo "Available targets:"
	@echo "  install          - Install system dependencies"
	@echo "  deploy           - Full deployment (frontend + backend + nginx)"
	@echo "  deploy-frontend  - Deploy frontend only"
	@echo "  deploy-backend   - Deploy backend only (specify BACKEND=go|py|kt|rs)"
	@echo "  clean            - Stop services and remove deployment files"
	@echo "  status           - Show service status"
	@echo "  logs             - Show service logs"
	@echo "  restart          - Restart all services"
	@echo ""
	@echo "Examples:"
	@echo "  make deploy DOMAIN=pebble.com EMAIL=admin@pebble.com BACKEND=go"
	@echo "  make deploy-backend BACKEND=py"
	@echo "  make status"

install:
	@echo "=== Installing system dependencies ==="
	@sudo $(SCRIPT_DIR)/install-deps.sh

deploy: install deploy-frontend deploy-backend
	@echo "=== Configuring Nginx and SSL ==="
	@sudo DOMAIN=$(DOMAIN) EMAIL=$(EMAIL) $(SCRIPT_DIR)/setup-nginx.sh
	@echo ""
	@echo "=== Deployment completed successfully ==="
	@echo "Frontend: https://$(DOMAIN)"
	@echo "Backend API: https://$(DOMAIN)/api"
	@echo ""
	@echo "Use 'make status' to check service status"
	@echo "Use 'make logs' to view service logs"

deploy-frontend:
	@echo "=== Deploying frontend ==="
	@sudo $(SCRIPT_DIR)/deploy-frontend.sh

deploy-backend:
	@echo "=== Deploying backend: $(BACKEND) ==="
	@sudo BACKEND=$(BACKEND) $(SCRIPT_DIR)/deploy-backend.sh

clean:
	@echo "=== Cleaning up deployment ==="
	@sudo $(SCRIPT_DIR)/cleanup.sh

status:
	@echo "=== Service Status ==="
	@sudo systemctl status pebble-backend --no-pager || true
	@echo ""
	@sudo systemctl status nginx --no-pager || true

logs:
	@echo "=== Backend Logs (last 50 lines) ==="
	@sudo journalctl -u pebble-backend -n 50 --no-pager
	@echo ""
	@echo "=== Nginx Logs (last 20 lines) ==="
	@sudo tail -n 20 /var/log/nginx/error.log

restart:
	@echo "=== Restarting services ==="
	@sudo systemctl restart pebble-backend
	@sudo systemctl reload nginx
	@echo "Services restarted successfully"