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
│   ├── embed.go           # Embeds frontend assets
│   └── dist/              # Frontend build output embedded in the Go binary
├── docker/
└── data/
```

## Runtime Behavior

- The server reads runtime config from `./data/config.yaml`
- If the file does not exist, a default config is generated automatically
- Frontend assets in `static/dist/` are embedded in the Go binary and do not need to be deployed separately
- Images are stored under `./data/img`
- An empty database is initialized with a default `Hello World` article

## Quick Start

### Requirements

- Go 1.24+
- Node.js 20.19+ or 22.12+ for frontend or plugin builds
- SQLite for the simplest setup
- MySQL with ngram parser for better full-text search
- PostgreSQL with zhparser for Chinese full-text search

### 1. Build and run

The full build creates the frontend output, syncs it to `static/dist/`, and then
embeds it in a single Go executable:

```bash
make build
./build/bin/mkBlog
```

If you already placed the complete frontend `dist/` directory under `static/`
(at `static/dist/`), run `make backend-build` directly.

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

After building, place the complete `frontend/dist/` directory under `static/`,
then rebuild the Go executable:

```bash
make sync-static backend-build
```

The resulting `build/bin/mkBlog` contains all frontend assets.

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
