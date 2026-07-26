**English Version: [English](README_en.md)**

# mkBlog

一个基于 Go + Gin + GORM + Vue 3 的极简个人博客系统。前后端统一部署，默认使用 SQLite 即可启动，也支持 MySQL / PostgreSQL。项目同时提供 VS Code 和 Obsidian 上传插件，适合以 Markdown 为中心的写作流。

[预览](https://mkitsdts.top:8080)

![Home](./docs/images/home.png)
![Article](./docs/images/article.png)
![ApplyFriend](./docs/images/apply_friend.png)

## 功能特性

- Markdown 文章发布、更新、删除
- 分类筛选与分页展示
- 文章搜索
- 文章评论开关
- 友链展示与申请
- 图片上传并自动转换为 WebP
- 基于 Bearer Token 的写接口鉴权
- 限流、黑名单、Bloom Filter
- TLS / HTTP3 / 自动证书控制
- VS Code / Obsidian 插件上传工作流

## 技术栈

### 后端

- Go 1.24
- Gin
- GORM
- SQLite / MySQL / PostgreSQL

### 前端

- Vue 3
- Vite
- Element Plus
- Axios
- markdown-it
- highlight.js

## 项目结构

```text
.
├── main.go
├── service/               # API 与服务启动
├── pkg/                   # 中间件、缓存、TLS、数据库等
├── frontend/              # Vue 前端
├── plugin/vscode/         # VS Code 上传插件
├── plugin/obsidian/       # Obsidian 上传插件
├── static/                # 前端构建产物
├── docker/                # Dockerfile
└── data/                  # 运行时数据目录（首次启动自动生成）
```

## 运行机制

- 后端启动后会读取 `./data/config.yaml`
- 如果配置文件不存在，程序会自动生成默认配置
- 前端构建产物由后端统一静态托管
- 图片默认保存在 `./data/static/images`
- 空数据库会自动插入一篇 `Hello World` 示例文章

## 快速开始

### 环境要求

- Go 1.24+
- Node.js 20.19+ 或 22.12+，仅构建前端或插件时需要
- SQLite 可直接使用
- MySQL 需要 ngram parser 才能获得更好的全文搜索效果
- PostgreSQL 中文搜索需要 zhparser 扩展

### 1. 构建并启动后端

```bash
go build -o mkBlog .
./mkBlog
```

首次启动会自动生成：

```text
./data/config.yaml
./data/app.db
```

### 2. 修改配置

编辑 `./data/config.yaml`，重点关注以下字段：

- `database.kind`: `sqlite3` / `mysql` / `postgres`
- `database.host`: SQLite 下填写数据库文件名，MySQL / PostgreSQL 下填写主机地址
- `server.port`: 服务端口
- `server.imageSavePath`: 图片保存目录
- `server.devmode`: 是否启用开发模式
- `server.http3_enabled`: 是否启用 HTTP/3
- `tls.enabled`: 是否启用 HTTPS
- `auth.enabled`: 是否启用 Bearer Token 鉴权
- `auth.secret`: 写接口使用的 Token
- `site.signature`: 首页签名
- `site.about`: 关于页面内容
- `site.avatarPath`: 头像文件名
- `site.server`: 对外访问地址
- `site.comment_enabled`: 是否开启评论
- `site.icp`: 备案号

### 3. 构建前端

```bash
cd frontend
npm install
npm run build
```

前端构建完成后，需要把产物复制到仓库根目录的 `static/` 下供后端托管。

### 4. 访问项目

默认访问地址：

```text
http://127.0.0.1:4801
```

如果正确设置了 `site.server`，前端会使用该地址作为生产环境 API 根地址。

## 数据库说明

### SQLite

- 默认方案
- 无需额外安装数据库
- 搜索功能使用 LIKE 回退方案

### MySQL

- 适合正式部署
- 搜索依赖 FULLTEXT + ngram parser

### PostgreSQL

- 支持中文搜索
- 需要安装 zhparser 扩展

## API 概览

### 公开接口

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

### 管理接口

- `PUT /api/article/:title`
- `PUT /api/image`
- `DELETE /api/article/:title`
- `POST /api/blockip`

当 `auth.enabled: true` 时，上述管理接口需要携带：

```http
Authorization: Bearer <your-token>
```

## 图片与 Markdown 约定

- 图片按文章标题归档保存
- 非 WebP 图片上传后会自动转为 `.webp`
- 文内引用图片时可以不写扩展名
- 如果手动写扩展名，建议写 `.webp`
- 插件默认会上传 Markdown 文件以及同名文件夹中的图片资源

## 插件

### VS Code

见 [plugin/vscode/README.md](plugin/vscode/README.md)

### Obsidian

见 [plugin/obsidian/README.md](plugin/obsidian/README.md)

## Docker

直接运行由 GitHub Actions 发布的多架构镜像：

```bash
docker pull ghcr.io/mkitsdts/mkblog:latest
docker run -d --name mkblog \
  -p 4801:4801 \
  -v /etc/mkblog:/etc/mkblog \
  ghcr.io/mkitsdts/mkblog:latest
```

也可以从源码本地构建：

```bash
docker build -f docker/Dockerfile -t mkblog:latest .
```

容器使用 `/etc/mkblog` 保存配置、数据库、日志和上传图片。请将该目录挂载到宿主机以持久化运行数据。

GHCR 首次创建的软件包默认为私有。如果需要免登录拉取镜像，请在 GitHub 的软件包设置中将 `mkblog` 的可见性改为 `Public`。

## CI/CD

- `main` 分支更新时，只有代码相关目录、Docker 文件或 GitHub Actions 工作流发生变更，才会触发服务器部署
- 部署工作流会在服务器上执行 `make release`
- 后端、前端、静态资源、Dockerfile 或镜像工作流在 `main` 分支发生变化时，会自动构建并推送镜像
- 发布 GitHub Release 时也会构建镜像，并生成 `latest`、提交 SHA 和版本号标签
- 镜像发布到 GitHub Container Registry：`ghcr.io/mkitsdts/mkblog`
- Actions 页面支持通过 `workflow_dispatch` 手动触发镜像构建
- 镜像同时支持 `linux/amd64` 与 `linux/arm64`

## 说明

- 前后端同源部署，默认不存在 CORS 问题
- 如需 HTTPS，可启用 `tls.enabled`
- 如需自动签发证书，可配置 `cert_control`
- VS Code 和 Obsidian 插件都支持通过配置 Base URL 连接远端博客服务

## License

[MIT](LICENSE)
