# mkBlog
## 介绍

本项目是一个基于 Go 语言的 Gin 框架搭建的个人博客后端及后台管理，支持 HTTP1.1, HTTPS 协议，主要用途是记录内容积累

## 技术栈

整个项目只使用到 MySQL

使用到的包有：

* slog                                  日志库

* gin                                   Web框架

* gorm                                  ORM框架

## 状态码

* 200 正常访问

* 400 非法参数

* 404 访问资源不存在

* 500 服务器内部出错

## 启动方式

配置好 go 环境和 MySQL 环境，有条件的可以使用 docker 部署

真机部署：
* 1、直接运行 go run main.go

* 2、编译运行 go build -o blog 然后 ./blog

docker部署：

// ...existing code...
docker部署：

* 1、构建Docker镜像：
  ```bash
  docker build -t [your_name] .
  ```

* 2、运行Docker容器：
  ```bash
  docker run -p [host_port]:[container_port] [your_name]
  ```

* 3、如果需要挂载数据卷或传递环境变量：
  ```bash
  docker run -p 8080:8080 -v $(pwd)/resource:/app/resource -e DB_HOST=host.docker.internal [your_name]
  ```

示例：
```bash
# 构建镜像
docker build -t mkblog-backend .

# 启动容器（映射到本机8080端口）
docker run -p 8080:8080 mkblog-backend

# 后台运行
docker run -d -p 8080:8080 --name mkblog mkblog-backend
```
// ...existing code... 

## 启动参数

create + 文件名:  创建一个名为 title.md 并且创建一个名为 title 的文件夹