# mkBlog

## 介绍

本项目是一个基于Go语言的Gin框架搭建的后端博客，支持 HTTP1.1, HTTP2.0, HTTPS 协议，主要用途是写一些东西记录

## 技术栈

整个项目使用到的中间件有Redis，数据库是MySQL

Redis用于缓存文章内容，RabbitMQ用于缓存用户请求，MySQL用于存储文章内容

使用到的标准库包有：

* slog          日志库



使用到的第三方包有：

* gin           Web框架

* gorm          连接MySQL

* jwt           生成token 

* md5           生成md5

* viper         读取json、md文件

* fsnotify      监听项目文件的变化

* opentracing   追踪数据

* validator     进行数据验证

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

        - package

        |

        |
        
        - post

        |

        |

        - client

## 状态码

* 200 正常访问

* 403 无权限访问

* 404 访问资源不存在

* 500 服务器内部出错