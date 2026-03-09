**Chinese Version: [中文](README.md)**

# mkBlog

mkBlog is a minimalist personal blog system built with Go, Gin, GORM, Vue 3, and Vite. It serves both frontend and backend from one deployment target, uses SQLite by default, and also supports MySQL and PostgreSQL. VS Code and Obsidian uploader plugins are included for a Markdown-first workflow.

[Preview](https://mkitsdts.top:8080)

![Home](./docs/images/home.png)
![Article](./docs/images/article.png)
![ApplyFriend](./docs/images/apply_friend.png)

## Features

- Create, update, and delete Markdown articles
- Category filtering and pagination
- Article search
- Switchable comments
- Friend links and friend link applications
- Image upload with automatic WebP conversion
- Bearer Token protection for write APIs
- Rate limiting, blacklist mode, and Bloom filter
- TLS / HTTP3 / automatic certificate control
- VS Code / Obsidian uploader workflow

## Tech Stack

### Backend

- Go 1.24
- Gin
- GORM
- SQLite / MySQL / PostgreSQL

### Frontend

- Vue 3
- Vite
- Element Plus
- Axios
- markdown-it
- highlight.js

## Project Layout

```text
.
├── main.go
├── service/
├── pkg/
├── frontend/
├── plugin/vscode/
├── plugin/obsidian/
├── static/
├── docker/
└── data/
```

## Runtime Behavior

- The server reads runtime config from `./data/config.yaml`
- If the file does not exist, a default config is generated automatically
- Frontend build assets are served by the Go backend
- Images are stored under `./data/static/images`
- An empty database is initialized with a default `Hello World` article

## Quick Start

### Requirements

- Go 1.24+
- Node.js 20.19+ or 22.12+ for frontend or plugin builds
- SQLite for the simplest setup
- MySQL with ngram parser for better full-text search
- PostgreSQL with zhparser for Chinese full-text search

### 1. Build and run the backend

```bash
go build -o mkBlog .
./mkBlog
```

On first start, mkBlog creates:

```text
./data/config.yaml
./data/app.db
```

### 2. Edit configuration

Update `./data/config.yaml` and review these fields:

- `database.kind`: `sqlite3` / `mysql` / `postgres`
- `database.host`: file name for SQLite, host address for MySQL / PostgreSQL
- `server.port`
- `server.imageSavePath`
- `server.devmode`
- `server.http3_enabled`
- `tls.enabled`
- `auth.enabled`
- `auth.secret`
- `site.signature`
- `site.about`
- `site.avatarPath`
- `site.server`
- `site.comment_enabled`
- `site.icp`

### 3. Build the frontend

```bash
cd frontend
npm install
npm run build
```

After building, copy the frontend output into the root `static/` directory so it can be served by the backend.

### 4. Open the site

Default address:

```text
http://127.0.0.1:4801
```

If `site.server` is configured correctly, the frontend uses that value as the production API root.

## Database Notes

### SQLite

- Default option
- No extra database service required
- Search falls back to LIKE queries

### MySQL

- Better suited for production deployments
- Search uses FULLTEXT + ngram parser

### PostgreSQL

- Supported
- Chinese search requires zhparser

## API Overview

### Public APIs

- `GET /api/site`
- `GET /api/articles`
- `GET /api/allarticles`
- `GET /api/article/:title`
- `GET /api/search`
- `GET /api/categories`
- `GET /api/friends`
- `POST /api/friends`
- `GET /api/comments`
- `POST /api/comments`

### Admin APIs

- `PUT /api/article/:title`
- `PUT /api/image`
- `DELETE /api/article/:title`
- `POST /api/blockip`

When `auth.enabled: true`, admin APIs require:

```http
Authorization: Bearer <your-token>
```

## Image and Markdown Conventions

- Images are grouped by article title
- Uploaded non-WebP images are converted to `.webp`
- In article content, image paths can omit the extension
- If an extension is written manually, use `.webp`
- The plugins upload Markdown files together with images in same-name folders

## Plugins

### VS Code

See [plugin/vscode/README.md](plugin/vscode/README.md)

### Obsidian

See [plugin/obsidian/README.md](plugin/obsidian/README.md)

## Docker

```bash
docker build -f docker/Dockerfile -t mkblog:latest .
docker run -d --name mkblog -p 4801:4801 -v /etc/mkblog:/app/data mkblog:latest
```

Runtime data inside the container is stored in `/app/data`.

## Online Install and Update

If you do not want to use Docker, you can install mkBlog locally and register it as a system service with the install script.

### Requirements

- Git
- Go 1.24+
- Node.js 20.19+ or 22.12+
- npm
- make
- `systemd` on Linux
- `launchd` on macOS

### Install

```bash
curl -fsSL https://raw.githubusercontent.com/mkitsdts/mkBlog/main/scripts/install.sh | bash -s -- install
```

### Update

```bash
curl -fsSL https://raw.githubusercontent.com/mkitsdts/mkBlog/main/scripts/install.sh | bash -s -- update
```

### Common Commands

```bash
curl -fsSL https://raw.githubusercontent.com/mkitsdts/mkBlog/main/scripts/install.sh | bash -s -- start
curl -fsSL https://raw.githubusercontent.com/mkitsdts/mkBlog/main/scripts/install.sh | bash -s -- stop
curl -fsSL https://raw.githubusercontent.com/mkitsdts/mkBlog/main/scripts/install.sh | bash -s -- restart
curl -fsSL https://raw.githubusercontent.com/mkitsdts/mkBlog/main/scripts/install.sh | bash -s -- status
curl -fsSL https://raw.githubusercontent.com/mkitsdts/mkBlog/main/scripts/install.sh | bash -s -- uninstall
```

### Optional Environment Variables

- `MKBLOG_INSTALL_DIR`: install directory, default `~/.local/share/mkblog`
- `MKBLOG_REPO_URL`: repository URL
- `MKBLOG_REPO_REF`: branch or tag, default `main`
- `MKBLOG_SERVICE_NAME`: systemd service name on Linux, default `mkblog`
- `MKBLOG_LAUNCHD_LABEL`: launchd label on macOS, default `com.mkblog.app`

## CI/CD

- Deployments from `main` only run when code-related directories, Docker files, or GitHub Actions workflows change
- The deploy workflow runs `make release` on the remote server
- Publishing a GitHub Release automatically builds and pushes multi-architecture Docker images for `linux/amd64` and `linux/arm64`
- Changes under `docker/**` or `.github/workflows/**` on `main` also trigger Docker image build and push

## Notes

- Frontend and backend are served from the same origin
- Enable `tls.enabled` if you want HTTPS
- Configure `cert_control` if you want automated certificate management
- Both VS Code and Obsidian plugins can connect to a remote mkBlog instance through Base URL settings

## License

[MIT](LICENSE)
