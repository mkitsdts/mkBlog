# mkBlog

 Go 开发的极简个人博客系统，专注于内容创作和技术积累。

## 项目介绍

mkBlog 是一个轻量级的个人博客系统，支持 Markdown 文章、文章分类。系统设计简洁，易于部署和维护。

## 技术栈

- **Go 1.24** - 主要编程语言
- **Gin** - Web 框架
- **GORM** - ORM 框架
- **MySQL** - 数据库

## 功能特性

- ✅ **文章管理** - 支持 Markdown 格式文章的创建、编辑和展示
- ✅ **文章分类** - 按分类组织文章内容
- ✅ **文章搜索** - 支持关键词搜索文章
- ✅ **分页显示** - 文章列表分页展示
- ✅ **友链管理** - 友链展示和申请功能
- ❌ **图片支持** - 文章内图片展示

## 快速开始

### 环境要求
- Go 1.24+
- MySQL 8.0+

### 真机部署

1. **配置数据库**
   ```bash
   # 创建数据库
   CREATE DATABASE mkblog CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
   ```

2. **配置文件**
   
   编辑 `backend/config.json`：
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

3. **启动后端服务**
```bash
go mod tidy
go run main.go
```

### Docker 部署

**数据库 Docker 部署：**
```bash
cd docker
docker-compose up -d
```

## 使用说明

修改头像和个性签名目前只能在 frontend/src 文件夹下修改 config.js 文件

目前暂时没有很好的方式上传文章，正在想一个方案

## 访问地址

前后端统一：http://localhost:8080

如果需要配置 TLS 证书或修改端口，目前只能自行修改代码

## 开发进度

- [x] 基础文章系统
- [x] 文章分类和搜索
- [x] 友链管理
- [x] 响应式 UI
- [ ] 后台管理界面
- [ ] 评论系统
- [ ] RSS 订阅
- [ ] SEO 优化