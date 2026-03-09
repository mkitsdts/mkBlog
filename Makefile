.PHONY: help build backend-build frontend-build sync-static release-build release release-install release-start release-stop release-restart release-uninstall dev-backend dev-frontend clean docker-build

APP_NAME := mkBlog
ROOT_DIR := $(abspath .)
BUILD_DIR := $(ROOT_DIR)/build
BIN_DIR := $(BUILD_DIR)/bin
BIN_PATH := $(BIN_DIR)/$(APP_NAME)
FRONTEND_DIR := $(ROOT_DIR)/frontend
FRONTEND_DIST := $(FRONTEND_DIR)/dist
STATIC_DIR := $(ROOT_DIR)/static
STATIC_ASSETS_DIR := $(STATIC_DIR)/assets
DATA_DIR := $(ROOT_DIR)/data

OS := $(shell uname -s)
SERVICE_NAME ?= mkblog
SYSTEMD_SERVICE_PATH ?= /etc/systemd/system/$(SERVICE_NAME).service
LAUNCHD_LABEL ?= com.mkblog.app
LAUNCHD_PLIST_PATH ?= $(HOME)/Library/LaunchAgents/$(LAUNCHD_LABEL).plist

help:
	@echo "Available targets:"
	@echo "  build            Build backend, frontend, and sync static assets"
	@echo "  backend-build    Build Go binary to build/bin/$(APP_NAME)"
	@echo "  frontend-build   Install frontend deps and build Vite assets"
	@echo "  sync-static      Copy frontend build output into static/"
	@echo "  release          One-command release build + service install + start"
	@echo "  release-install  Install service definition for Linux(systemd) or macOS(launchd)"
	@echo "  release-start    Start registered service"
	@echo "  release-stop     Stop registered service"
	@echo "  release-restart  Restart registered service"
	@echo "  release-uninstall Remove registered service definition"
	@echo "  dev-backend      Run backend with go run"
	@echo "  dev-frontend     Run frontend with Vite"
	@echo "  clean            Remove local build artifacts"
	@echo "  docker-build     Build Docker image mkblog:latest"

build: release-build

backend-build:
	@echo "Building Go binary..."
	@mkdir -p "$(BIN_DIR)"
	go build -o "$(BIN_PATH)" .

frontend-build:
	@echo "Installing frontend dependencies and building..."
	cd "$(FRONTEND_DIR)" && npm ci && npm run build

sync-static:
	@echo "Syncing frontend assets to static/..."
	@test -d "$(FRONTEND_DIST)" || (echo "Missing frontend build output at $(FRONTEND_DIST). Run 'make frontend-build' first." && exit 1)
	@mkdir -p "$(STATIC_DIR)" "$(STATIC_ASSETS_DIR)"
	@rm -f "$(STATIC_DIR)/index.html"
	@rm -rf "$(STATIC_ASSETS_DIR)"
	@mkdir -p "$(STATIC_ASSETS_DIR)"
	cp "$(FRONTEND_DIST)/index.html" "$(STATIC_DIR)/index.html"
	cp -R "$(FRONTEND_DIST)/assets/." "$(STATIC_ASSETS_DIR)/"

release-build: backend-build frontend-build sync-static
	@echo "Release build completed."

release: release-build release-install release-start
	@echo "Release workflow completed."

release-install: release-build
ifeq ($(OS),Linux)
	@echo "Installing systemd service to $(SYSTEMD_SERVICE_PATH)..."
	@mkdir -p "$(BUILD_DIR)"
	@sed \
		-e 's|__APP_NAME__|$(APP_NAME)|g' \
		-e 's|__ROOT_DIR__|$(ROOT_DIR)|g' \
		-e 's|__BIN_PATH__|$(BIN_PATH)|g' \
		-e 's|__USER__|$(shell id -un)|g' \
		-e 's|__GROUP__|$(shell id -gn)|g' \
		"$(ROOT_DIR)/deploy/systemd/mkblog.service.tpl" > "$(BUILD_DIR)/$(SERVICE_NAME).service"
	sudo install -m 0644 "$(BUILD_DIR)/$(SERVICE_NAME).service" "$(SYSTEMD_SERVICE_PATH)"
	sudo systemctl daemon-reload
	sudo systemctl enable "$(SERVICE_NAME)"
else ifeq ($(OS),Darwin)
	@echo "Installing launchd plist to $(LAUNCHD_PLIST_PATH)..."
	@mkdir -p "$(BUILD_DIR)" "$(HOME)/Library/LaunchAgents" "$(DATA_DIR)"
	@sed \
		-e 's|__LABEL__|$(LAUNCHD_LABEL)|g' \
		-e 's|__ROOT_DIR__|$(ROOT_DIR)|g' \
		-e 's|__BIN_PATH__|$(BIN_PATH)|g' \
		-e 's|__STDOUT_PATH__|$(DATA_DIR)/launchd.stdout.log|g' \
		-e 's|__STDERR_PATH__|$(DATA_DIR)/launchd.stderr.log|g' \
		"$(ROOT_DIR)/deploy/launchd/com.mkblog.app.plist.tpl" > "$(BUILD_DIR)/$(LAUNCHD_LABEL).plist"
	cp "$(BUILD_DIR)/$(LAUNCHD_LABEL).plist" "$(LAUNCHD_PLIST_PATH)"
	launchctl bootout "gui/$$(id -u)" "$(LAUNCHD_PLIST_PATH)" >/dev/null 2>&1 || true
	launchctl bootstrap "gui/$$(id -u)" "$(LAUNCHD_PLIST_PATH)"
else
	@echo "Unsupported OS: $(OS)" && exit 1
endif

release-start:
ifeq ($(OS),Linux)
	sudo systemctl start "$(SERVICE_NAME)"
	sudo systemctl status "$(SERVICE_NAME)" --no-pager || true
else ifeq ($(OS),Darwin)
	launchctl bootstrap "gui/$$(id -u)" "$(LAUNCHD_PLIST_PATH)" >/dev/null 2>&1 || true
	launchctl kickstart -k "gui/$$(id -u)/$(LAUNCHD_LABEL)"
	launchctl print "gui/$$(id -u)/$(LAUNCHD_LABEL)" || true
else
	@echo "Unsupported OS: $(OS)" && exit 1
endif

release-stop:
ifeq ($(OS),Linux)
	sudo systemctl stop "$(SERVICE_NAME)"
else ifeq ($(OS),Darwin)
	launchctl bootout "gui/$$(id -u)" "$(LAUNCHD_PLIST_PATH)" >/dev/null 2>&1 || launchctl unload "$(LAUNCHD_PLIST_PATH)" >/dev/null 2>&1 || true
else
	@echo "Unsupported OS: $(OS)" && exit 1
endif

release-restart:
ifeq ($(OS),Linux)
	sudo systemctl restart "$(SERVICE_NAME)"
	sudo systemctl status "$(SERVICE_NAME)" --no-pager || true
else ifeq ($(OS),Darwin)
	launchctl bootout "gui/$$(id -u)" "$(LAUNCHD_PLIST_PATH)" >/dev/null 2>&1 || true
	launchctl bootstrap "gui/$$(id -u)" "$(LAUNCHD_PLIST_PATH)"
	launchctl kickstart -k "gui/$$(id -u)/$(LAUNCHD_LABEL)"
	launchctl print "gui/$$(id -u)/$(LAUNCHD_LABEL)" || true
else
	@echo "Unsupported OS: $(OS)" && exit 1
endif

release-uninstall:
ifeq ($(OS),Linux)
	-sudo systemctl stop "$(SERVICE_NAME)"
	-sudo systemctl disable "$(SERVICE_NAME)"
	-sudo rm -f "$(SYSTEMD_SERVICE_PATH)"
	sudo systemctl daemon-reload
else ifeq ($(OS),Darwin)
	-launchctl bootout "gui/$$(id -u)" "$(LAUNCHD_PLIST_PATH)" >/dev/null 2>&1 || launchctl unload "$(LAUNCHD_PLIST_PATH)" >/dev/null 2>&1 || true
	-rm -f "$(LAUNCHD_PLIST_PATH)"
else
	@echo "Unsupported OS: $(OS)" && exit 1
endif

dev-backend:
	go run . -debug

dev-frontend:
	cd "$(FRONTEND_DIR)" && npm run dev

clean:
	@echo "Removing build artifacts..."
	@rm -rf "$(BUILD_DIR)" "$(FRONTEND_DIST)" "$(FRONTEND_DIR)/build"
	@rm -f "$(ROOT_DIR)/$(APP_NAME)"

docker-build:
	docker build -f docker/Dockerfile -t mkblog:latest .
