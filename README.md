# mkBlog
目前还是玩具，不具备实用性
## 介绍

本项目是一个基于 Go 语言的 Gin 框架搭建的个人博客后端及后台管理，支持 HTTP1.1, HTTPS 协议，主要用途是记录内容积累

## 技术栈

整个项目只使用到 Redis

使用到的包有：

* slog                                  日志库

* gin                                   Web框架

* github.com/redis/go-redis/v9          连接MySQL

## 状态码

* 200 正常访问

* 404 访问资源不存在

* 500 服务器内部出错

## 启动方式

配置好 go 环境和 redis 环境，有条件的可以使用 docker 部署

真机部署：
* 1、直接运行 go run main.go

* 2、编译运行 go build -o blog 然后 ./blog

docker部署：
* 

## 启动参数

create + 文件名:  创建一个名为 title.md 并且创建一个名为 title 的文件夹

update: 更新数据库