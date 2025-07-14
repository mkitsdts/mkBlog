# mkBlog

一个基于 Go + Vue.js 开发的极简个人博客系统，专注于内容创作和技术积累。

## 项目介绍

mkBlog 是一个轻量级的个人博客系统，支持 Markdown 文章编写、文章分类。系统设计简洁，易于部署和维护。

## 技术栈

### 后端
- **Go 1.24** - 主要编程语言
- **Gin** - Web 框架
- **GORM** - ORM 框架
- **MySQL** - 数据库

### 前端
- **Vue 3** - 前端框架
- **Vue Router** - 路由管理
- **Vite** - 构建工具
- **Marked** - Markdown 解析器

## 功能特性

- ✅ **文章管理** - 支持 Markdown 格式文章的创建、编辑和展示
- ✅ **文章分类** - 按分类组织文章内容
- ✅ **文章搜索** - 支持关键词搜索文章
- ✅ **分页显示** - 文章列表分页展示
- ❌ **友链管理** - 友链展示和申请功能
- ❌ **图片支持** - 文章内图片展示

## 项目结构

```
mkBlog/
├── backend/                # 后端代码
│   ├── config/             # 配置管理
│   ├── models/             # 数据模型
│   ├── pkg/                # 基础包
│   ├── service/            # 业务逻辑
│   ├── resource/           # Markdown 文章存储
│   ├── main.go             # 程序入口
│   └── config.json         # 配置文件
├── frontend/               # 前端代码
│   ├── src/
│   │   ├── components/     # Vue 组件
│   │   ├── router/         # 路由配置
│   │   └── assets/         # 静态资源
│   ├── index.html
│   └── package.json
└── README.md
```

## 快速开始

### 环境要求
- Go 1.24+
- Node.js 16+
- MySQL 8.0+

### 后端部署

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
           "user": "root",
           "password": "your_password",
           "name": "mkblog"
       }
   }
   ```

3. **启动后端服务**
   ```bash
   cd backend
   go mod tidy
   go run main.go
   ```

### 前端部署

1. **安装依赖**
   ```bash
   cd frontend
   npm install
   ```

2. **开发模式**
   ```bash
   npm run dev
   ```

3. **生产构建**
   ```bash
   npm run build
   npm run preview
   ```

### Docker 部署

**后端 Docker 部署：**
```bash
cd backend
docker build -t mkblog-backend .
docker run -p 8080:8080 mkblog-backend
```

## 使用说明

### 创建文章
```bash
# 在 backend 目录下执行
go run main.go create "你的文章标题"
# 最好用编译好的二进制文件
./mkblog create "你的文章标题"
```

这会在 `resource/` 目录下创建：
- `文章标题.md` - Markdown 文件
- `文章标题/` - 图片资源目录

### 文章格式
```markdown
---
title: 文章标题
created_time: 2024-01-01 12:00:00
updated_time: 2024-01-01 12:00:00
tags: 标签1,标签2
category: 分类名
author: 作者名
---

这里是文章内容...
```

必须按照文章格式书写，不然会解析错误

### 更新文章
```bash
go run main.go update
```

## API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/home?page=1` | 获取文章列表 |
| GET | `/articles/:title` | 获取文章详情 |
| GET | `/search?keyword=&page=1` | 搜索文章 |
| GET | `/images/:title/:filename` | 获取文章图片 |
| GET | `/friend` | 获取友链列表 |
| POST | `/friend/apply` | 申请友链 |

## 访问地址

- 前端访问：http://localhost:5173
- 后端 API：http://localhost:8080

## 开发进度

- [x] 基础文章系统
- [x] 文章分类和搜索
- [x] 友链管理
- [x] 响应式 UI
- [ ] 后台管理界面
- [ ] 评论系统
- [ ] RSS 订阅
- [ ] SEO 优化

## 许可证

本项目基于 [MIT License](backend/LICENSE) 开源。

---

💡 **提示**：这是一个玩具项目，适合 Go 和 Vue.js 初学者参考学习。