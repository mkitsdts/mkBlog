**Chinese Version: [中文](README.md)**

# mkBlog

A minimalist personal blog system written in Go, focused on content creation and technical notes, with one-click deployment for both frontend and backend.

[Preview](https://mkitsdts.top:8080)

![Home](./docs/images/home.png)

![Article](./docs/images/article.png)

![ApplyFriend](./docs/images/apply_friend.png)

## Project Overview

mkBlog is a lightweight personal blog system that supports Markdown articles and article categories. The system is designed to be simple, easy to deploy, and easy to maintain.

## Usage

You can change the avatar, signature, and personal introduction in the configuration file [config.yaml] by editing the corresponding values under site.

When writing posts with images, you don't need to include the image file extension in the path. If you do include an extension, please use webp, because the backend converts uploaded images to webp format.

Currently, file uploads are handled via a VS Code extension that acts as a simple admin backend. By configuring the extension with the correct address you can create, update, and delete blog content. The extension is in the plugin folder and can be built by yourself or you can use the distributed version.

## Technology Stack

- Go 1.24 — primary programming language  
- Gin — web framework  
- GORM — ORM  
- MySQL — database

## Features

- ✅ Article management — create, edit, and display Markdown articles  
- ✅ Article categories — organize content by category  
- ✅ Article search — keyword search for articles  
- ✅ Pagination — paginated article lists  
- ✅ Friend links — display and request friend links  
- ✅ Image support — display images inside articles  
- ✅ Article comments — comment and display comments under articles

New: A VS Code plugin for blog management is included in the plugin folder. It currently supports creating and deleting articles and images, and parsing tags from MD files. Article checking and formatting are not yet complete, so the extension is not published. Future plans include a more complete backend management experience in VS Code (create/upload/delete images and articles, and some format checks).

A Bloom filter, rate limiter, and blacklist mode have been added.

## Quick Start

### Requirements
- Go 1.24+  
- MySQL 8.0+ (ngram tokenizer required)

### Deploy on a real machine

1. Configure the database:
   ```sql
   -- Create database
   CREATE DATABASE mkblog CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   ```

2. Configuration file

   Edit `config.yaml`. See comments in [config.yaml] for details.

3. Start the service

The simplest way is to use the Makefile:

```bash
make all
```

You can also use GitHub Actions for automated deployment (requires setting secrets in GitHub), or build a binary with `go build` and deploy the executable.

### Docker Deployment

**Database Docker deployment:**  
If using Docker, ensure environment variables are set correctly:
Set EV in docker-compose.yaml

```bash
cd docker
docker-compose up -d
```

## Access

Frontend and backend are served under the same address, so there are no CORS issues.

If you need TLS, enable the TLS option in `config.yaml` and place the TLS certificates in the static folder.