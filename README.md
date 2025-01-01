# mkBlog

## 介绍

本项目是一个基于Go语言的Gin框架搭建的后端博客，支持 HTTP1.1, HTTP2.0, HTTPS 协议，主要用途是写一些东西记录

## 技术栈

整个项目使用到的中间件有Redis、RabbitMQ，数据库是MySQL

Redis用于缓存文章内容，RabbitMQ用于缓存用户请求，MySQL用于存储文章内容

使用到的第三方包有：

* gin           Web框架

* gorm          连接MySQL

* jwt           生成token 

* md5           生成md5

* amqp091-go    连接RabbitMQ

* viper         读取json、md文件

* fsnotify      监听项目文件的变化

* opentracing   追踪数据

* validator     进行数据验证

## 流程

通过URL链接请求访问 -> 请求达到RabbitMQ -> 服务器从RabbitMQ获取请求 -> 检索redis -> 从redis中直接返回内容或从MySQL中获取内容后再返回

请求中包含了机器标识，如果相同标识在一个小时内访问，将会使用浏览器的缓存

## 架构

根据流程可以设计一个架构，大致内容如下：



## 状态码

* 200 正常访问

* 403 无权限访问

* 404 访问资源不存在

* 500 服务器内部出错