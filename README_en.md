**中文版本: [中文](README.md).**

# mkBlog

An ultra-minimal personal blog system written by Go, focused on content creation and technical accumulation. One‑command deploy for both backend and frontend.

![Home](./docs/images/home.png)

## Overview

mkBlog is a lightweight personal blogging platform supporting Markdown posts and category organization. It is intentionally simple, easy to deploy, and easy to maintain.

## Tech Stack

- **Go 1.24** – Primary language
- **Gin** – Web framework
- **GORM** – ORM
- **MySQL** – Database

## Features

- ✅ **Post Management** – Create, store and render Markdown articles
- ✅ **Categories** – Organize posts by category (multi-select filter on the homepage)
- ✅ **Search (keyword)** – Basic keyword searching (planned / partial depending on backend implementation)
- ✅ **Pagination** – Paged article list
- ✅ **Friend Links** – Display & apply for friendship links
- ❌ **Image Upload UI** – Not yet implemented (images can be referenced manually if hosted)

## Quick Start

### Requirements
- Go 1.24+
- MySQL 8.0+

### Native Deployment

1. **Create database**
    ```bash
    CREATE DATABASE mkblog CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
    ```

2. **Configure `config.json`** (at project root):
    ```json
    {
      "mysql": {
         "host": "localhost",
         "port": "3306",
         "user": "YOUR_USER",
         "password": "YOUR_PASSWORD",
         "name": "mkblog"
      }
    }
    ```

3. **Run backend**
    ```bash
    go mod tidy
    go run main.go
    ```

The server listens on `:8080` by default and serves both API and built frontend.

### Docker Deployment

From project root (compose file located in `docker/`):
```bash
docker compose -f docker/docker-compose.yaml up -d --build
```

Services:
- `db` – MySQL 8.0 (data persisted in named volume `db_data`)
- `app` – Go application (listens on port 8080)

## Usage Notes

- Avatar & signature: currently edited in `frontend/src/config.js` (hard‑coded for now).
- Article upload: no admin UI yet; insert via SQL or extend backend.
- Default “Hello World” article is auto-created when the database is empty.

## Access

Unified access (API + frontend SPA): `http://localhost:8080`

To enable TLS or change port, adjust the code (or wrap with a reverse proxy like Nginx / Caddy).

## Roadmap

- [x] Core article system
- [x] Category filtering (multi-select)
- [x] Friend links
- [x] Responsive UI & improved card design
- [x] Markdown rendering + syntax highlight
- [ ] Admin dashboard
- [ ] Comment system
- [ ] RSS feed
- [ ] SEO enhancements
- [ ] Image upload pipeline

## Project Structure (Simplified)

```
mkBlog
├── main.go              # Entry point
├── config/              # Configuration loader
├── service/             # HTTP handlers (articles, friends, categories)
├── models/              # GORM models
├── pkg/                 # Router & database setup
├── static/              # Built frontend assets (served in production)
└── frontend/            # Vite + Vue 3 source
```

## API (Selected)

| Method | Path                     | Description                     |
|--------|--------------------------|---------------------------------|
| GET    | /api/articles            | List articles (supports pagination & categories) |
| GET    | /api/article/:title      | Get article detail              |
| GET    | /api/categories          | Distinct category list          |
| GET    | /api/friends             | Friend links list               |
| POST   | /api/friends             | Apply for friend link           |
| PUT    | /api/article/:title      | (Prototype) add article         |

## Contributing

Issues & PRs are welcome. Keep code small, cohesive, and dependency-light.

## License

MIT

---
Feel free to adapt this project into your own writing platform. Happy blogging!