---
title: Connect to MySql
date: 2024-10-30 12:39:47
tags: Database
author: mkitsdts
---
在写一个数据库之前，借鉴mysql是很有必要的。所以写下这篇blog记录折腾MySql的过程。
* Ubuntu 22.04 LTS 系统环境
* 8.0.39-0ubuntu0.22.04.1 (Ubuntu) MySql版本

## 进入命令行
首先使用下列命令进入mysql命令行
```bash
mysql -uroot -p;
```

## 检查状态
然后使用下列命令检查mysql运行状态
```bash
sudo systemctl status mysql
```
若为绿色说明运行成功
否则mysql出错

## 原始文件
先进入最原始的数据文件夹，看下没有创建数据库的时候 mysql数据文件是什么样子
输入下列命令进入
```bash
cd /var/lib/mysql
```
可以看到文件夹里有如下几种文件：
'#ib_16384_0.dblwr'   binlog.000001   ca.pem            ibdata1        performance_schema   sys     
'#ib_16384_1.dblwr'   binlog.000002   client-cert.pem   ibtmp1         private_key.pem      undo_001
'#innodb_redo'        binlog.000003   client-key.pem    mkitsdts.pid   public_key.pem       undo_002
'#innodb_temp'        binlog.index    debian-5.7.flag   mysql          server-cert.pem      
 auto.cnf             ca-key.pem      ib_buffer_pool    mysql.ibd      server-key.pem          

#ib_16384_0.dblwr 和 #ib_16384_1.dblwr：
这些是InnoDB存储引擎的双写缓冲区（Doublewrite Buffer）文件。双写缓冲区是InnoDB用于处理页写操作的区域，以确保数据的持久性和一致性。

binlog.000001、binlog.000002、binlog.000003：
这些是二进制日志文件（binlog），记录了对数据库执行的所有更改（插入、更新、删除）。这些日志用于数据复制和恢复操作。

ca.pem、server-cert.pem、server-key.pem：
这些是SSL证书文件，用于MySQL服务器的身份验证和加密通信。ca.pem是证书颁发机构（CA）证书，server-cert.pem是服务器证书，server-key.pem是服务器私钥。

client-cert.pem、client-key.pem：
这些是客户端SSL证书和私钥文件，用于客户端身份验证。

ibdata1：
这是InnoDB存储引擎的系统表空间文件，包含了InnoDB的数据字典和系统表。

ibtmp1：
这是InnoDB的临时表空间文件，用于存储临时表和事务日志。

ib_buffer_pool：
这是InnoDB缓冲池的文件，用于缓存数据和索引，提高数据库性能。

mysql.ibd：
这是MySQL系统数据库mysql的表空间文件，包含了用户、权限、数据库等元数据。

performance_schema：
这是MySQL的性能模式数据库，用于监控和分析数据库服务器的性能。

sys：
这是MySQL的性能模式数据库中的一个系统表，用于存储性能模式的元数据。

undo_001、undo_002：
这些是InnoDB的回滚段文件，用于处理事务的回滚操作。

auto.cnf：
这是一个自动生成的配置文件，包含了数据目录的位置和其他一些自动配置的参数。

ca-key.pem：
这是证书颁发机构（CA）的私钥文件。

debian-5.7.flag：
这是一个标志文件，表明MySQL数据库已经初始化，通常在Debian或Ubuntu系统上看到。

private_key.pem、public_key.pem：
这些是私钥和公钥文件，用于加密和解密操作。

## 创建数据库
了解mysql基础文件后，创建一些数据库，然后看它究竟是怎么实现的。
**命令请一行一行输入**
```bash
CREATE DATABASE IF NOT EXISTS test;
CREATE DATABASE IF NOT EXISTS test1;
CREATE DATEBASE IF NOT EXISTS test2;
CREATE DATABASE IF NOT EXISTS test3;
CREATE DATEBASE IF NOT EXISTS test4;
CREATE DATABASE IF NOT EXISTS test5;
CREATE DATEBASE IF NOT EXISTS test6;
```
上面的命令执行后，mysql会自动创建数据库。
通过如下命令可查看创建的数据库
```bash
show databases;
```
然后可以看到如下输出
+--------------------+
| Database           |
+--------------------+
| information_schema |
| mysql              |
| performance_schema |
| sys                |
| test               |
| test1              |
| test2              |
| test3              |
| test4              |
| test5              |
| test6              |
+--------------------+
后面的test数据库就是刚刚用命令创建的啦！
回到数据文件夹，如下为数据文件夹在执行命令后的目录。
'#ib_16384_0.dblwr'   binlog.000001   ca.pem            ibdata1        performance_schema   sys     test4
'#ib_16384_1.dblwr'   binlog.000002   client-cert.pem   ibtmp1         private_key.pem      test    test5
'#innodb_redo'        binlog.000003   client-key.pem    mkitsdts.pid   public_key.pem       test1   test6
'#innodb_temp'        binlog.index    debian-5.7.flag   mysql          server-cert.pem      test2   undo_001
 auto.cnf             ca-key.pem      ib_buffer_pool    mysql.ibd      server-key.pem       test3   undo_002
 对比先前的目录，可以看到多了test,test1...先前创建的数据库。数据库都当成一个文件夹存储。因为还没有向数据库添加内容，所以这些文件夹目前都是空文件夹。
 
 ## 创建表
创建好数据库后，就可以向数据库添加表了。这里也写好了添加表的测试命令。
输入下列命令前先输入*USE test1*命令选择一个数据库
```bash
CREATE TABLE Users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    password VARCHAR(50) NOT NULL,
    email VARCHAR(100) DEFAULT NULL,
    age INT
);
```
执行完上面的命令后，可以输入如下命令查看表
```bash
show tables;
```
输入完命令，就可以看到刚刚创建的表啦！就像下面这样
+-----------------+
| Tables_in_test1 |
+-----------------+
| Users           |
| Users2          |
+-----------------+
再次回到数据文件夹，会看到数据文件夹多了一个Users.ibd的文件，这个就是刚刚创建的一个表。
假如再创建一个表，会发生什么呢？试试输入下面的命令
```bash
CREATE TABLE Users2 (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    password VARCHAR(50) NOT NULL,
    email VARCHAR(100) DEFAULT NULL,
    age INT
);
```
然后再次查看文件夹目录，此时会惊奇地发现多了一个Users2.ibd，所以我们可以大致推断，一个数据库就是一个文件夹，一个表就是文件夹里的一个文件，后缀为.ibd
以此类推，ibd里存储的就是数据文件。可能有人会好奇，ibd究竟存储的是什么呢？怀着一颗好奇的心，我输入了如下命令，也就是编辑Users.idb文件。
```bash
vim Users.ibd
```
然后发现，这是一串乱码。查了一下资料，了解到这不是简单的二进制文件，需要通过innodb存储引擎解码才能打开。
所以后续要使用mysql命令行对文件进行解析。

## 插入数据
创建好表后，可以向表插入数据。使用下面命令向表插入数据。
```bash
INSERT INTO Users (username, password, email, age)
VALUES ('testuser1', 'testpass1', 'test1@example.com', 25);
INSERT INTO Users (username, password, email, age)
VALUES ('testuser2', 'testpass2', 'test2@example.com', 30);
```
插入数据后，要查看插入的数据，可以通过以下命令实现
```bash
select * from Users;
```
然后会有如下输出：
+----+-----------+-----------+-------------------+------+
| id | username  | password  | email             | age  |
+----+-----------+-----------+-------------------+------+
|  1 | testuser1 | testpass1 | test1@example.com |   25 |
|  2 | testuser2 | testpass2 | test2@example.com |   30 |
+----+-----------+-----------+-------------------+------+
-----------------------
本次对mysql的探索到此结束，结合官网文档大概了解mysql的流程，有利于后续数据库的框架设计。