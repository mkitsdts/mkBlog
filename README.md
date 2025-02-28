# mkBlog

## 介绍

本项目是一个基于 Go 语言的 Gin 框架搭建的个人博客后端及后台管理，支持 HTTP1.1, HTTPS 协议，主要用途是记录内容积累

## 技术栈

整个项目使用到的中间件有 Redis ，数据库是 MySQL

Redis用于缓存文章内容， Kafka 用于缓存用户请求， MySQL 用于存储文章内容

使用到的包有：

* slog          日志库

* gin           Web框架

* gorm          连接MySQL

* redis         连接redis

* Kafka         连接mq

## 流程

通过URL链接请求访问 -> 服务器检索redis -> 从redis中直接返回内容或从MySQL中获取内容后再返回

请求中包含了机器标识，如果相同标识在一个小时内访问，将会使用浏览器的缓存

## 结构

根据流程设计文件架构，大致如下：

mkBlog  - internal

        |

        |

        - models

        |

        |

        - utils

## 状态码

* 200 正常访问

* 403 无权限访问

* 404 访问资源不存在

* 500 服务器内部出错