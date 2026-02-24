.PHONY: all build frontend-build copy run stop

all: stop build frontend-build copy run

build:
	@echo "Building Go binary..."
	go mod tidy
	go build -o mkBlog .

fbuild:
	@echo "Installing frontend deps and building..."
	rm -rf frontend/node_modules frontend/dist frontend/build frontend/package-lock.json
	cd frontend && npm install && npm run build

copy:
	@echo "Copying frontend build to static/ and static/assets/..."
	mkdir -p static static/assets
	@# detect build output dir (common: dist or build)
	if [ -d frontend/dist ]; then DIST=frontend/dist; elif [ -d frontend/build ]; then DIST=frontend/build; else echo "No frontend build dir (frontend/dist or frontend/build). Run 'make fbuild' first." && exit 1; fi; \
	cp -f "$$DIST/index.html" static/; \
	find "$$DIST" -type f \( -name '*.css' -o -name '*.js' \) -exec cp {} static/assets/ \;

run:
	@echo "Starting mkBlog..."
	./mkBlog & 1 & echo $$! > mkblog.pid
	@echo "mkBlog started, pid saved to mkblog.pid"

stop:
	@if [ -f mkblog.pid ]; then kill `cat mkblog.pid` && rm -f mkblog.pid && echo "Stopped mkBlog"; else echo "mkBlog not running (no mkblog.pid)"; fi
