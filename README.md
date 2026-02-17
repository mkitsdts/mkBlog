**English Version: [English](README_en.md)**

# mkBlog

 Go 开发的极简个人博客系统，专注于内容创作和技术积累，一键部署前后端。

 [预览](https://mkitsdts.top:8080)

![Home](./docs/images/home.png)

![Article](./docs/images/article.png)

![ApplyFriend](./docs/images/apply_friend.png)

## 项目介绍

mkBlog 是一个轻量级的个人博客系统，支持 Markdown 文章、文章分类。系统设计简洁，易于部署和维护。

## 使用说明

修改头像,个性签名和自我介绍可以在 [配置文件](config.yaml) 里进行，修改对应 site 下的值

编写博客时，如果有带图片，路径不需要填写后缀，如果填写后缀需要填写 webp ，因为后端接收图片时会将图片转换成 webp 格式。

## 管理工具

目前上传文件的方案是通过插件作为后台管理，填写地址实现对博客的增删改。已经存在的插件有 obsidian平台 和 vscode平台 [插件](plugin) 可以自行编译或者选择发行版的插件。

## 技术栈

- **Go 1.24** - 主要编程语言
- **Gin** - Web 框架
- **GORM** - ORM 框架
- **MySQL 或 Postgres 或 sqlite3** - 数据库

## 功能特性

- ✅ **文章管理** - 支持 Markdown 格式文章的创建、编辑和展示
- ✅ **文章分类** - 按分类组织文章内容
- ✅ **文章搜索** - 支持关键词搜索文章
- ✅ **分页显示** - 文章列表分页展示
- ✅ **友链管理** - 友链展示和申请功能
- ✅ **图片支持** - 文章内图片展示
- ✅ **文章评论** - 文章下评论及展示功能

便利功能： 

- 布隆过滤器
- 限流器
- 黑名单模式
- 自动 TLS 证书管理

## 快速开始

### 环境要求
- Go 1.24+
- （可选） MySQL 8.0+ （需要 ngram 分词器） 或 Postgres 18.0+ （需要 zhparser 插件）

支持 sqlite3 ，意味着不需要额外组件，直接就能启动运行。

### 真机部署

1. **配置数据库**
   
自行解决，默认使用 sqlite3 。不需要额外安装数据库

2. **配置文件**
   
在 `config.yaml` 里，前往 [配置文件](config.yaml) 查看注释

3. **启动服务**

最直接的方式是使用 Makefile ， 一行 make all 搞定

也可通过 github action 实现自动化部署。需要在 github 里配置秘密参数

或者 go build 编译成二进制文件后部署

### Docker 部署

**数据库 Docker 部署：**

自行解决 MySQL 和 Postgres 的需求。默认采用的 SQLite3 不需要额外部署数据库。

```bash
# 自动拉取，注意配置网络
docker pull mkitsdts/mkblog:latest
docker run -d --name mkblog -p 4801:4801 -v /etc/mkblog:/app/data mkitsdts/mkblog:latest
```

```bash
# 手动构建（比较耗时，建议优先用上面）
docker build -f docker/Dockerfile -t mkblog:latest . 
docker run -d --name mkblog -p 4801:4801 -v /etc/mkblog:/app/data mkblog:latest
```

通过docker部署后，配置文件会被写入到 /etc/mkblog目录下。

## 访问地址

前后端统一地址，所以没有跨域问题

建议使用 Let's Encrypt ，可以自动续费，一劳永逸。

如果需要配置 TLS 证书，可以在 config.yaml 里 tls 配置项的 enabled 选项打开，然后配置好自动管理 tls 证书。
