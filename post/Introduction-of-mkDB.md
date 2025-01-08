---
title: Introduction of mkDB
date: 2024-10-30 11:39:20
tags: mkDB
author: mkitsdts
---
mkDB是基于Linux系统开发的小型数据库，提供简易的C++接口，接口与SQL语句类似，后期考虑加上语法分析器。
mkDB支持多线程并发增删查改。
主要组件有基于B+树实现的存储引擎，支持网络链接的数据管理器，日志打印，数据恢复与备份。
## 系统功能

### 数据类型

int 

char

