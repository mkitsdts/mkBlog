---
title: Structure of mkDB
date: 2024-10-28 20:27:03
tags: mkDB
author: mkitsdts
--------
结构目录：include | test | docs 
            |       |

            |       |
                    _ _ _ | btree | page | table | log | index | database
            |

            |
            _ _ _ | general | page | table | log | index | database
                       |

                       |
                       _ _ _ | btree | comparer | define
-------------
mkDB存储引擎：
* 支持意外事务；
* 不支持外键；
* 缓存中只存在索引，不加载具体数据；
* 锁粒度为表级锁，故并发度差；
* 数据库发生意外，数据文件容易损坏；
综上，鉴定为玩具数据库！

## 索引
索引是帮助数据库高效获取数据的排好序的数据结构

包括：

* 二叉树，一种子节点最多为两个的树形数据结构
* 红黑树，一种从根节点到任意尾节点的路径之差不超过1的平衡二叉树
* 哈希表，一种通过哈希算法直接存储内存地址的数组
* B树，一种在节点存储多条索引元素以及附带数据的树形结构
* B+树，B树的变种，冗余存储了索引元素，在叶子节点存储了所有索引元素和附带数据，且含有双向指针

这里选取的是B+树作为索引，后面会探讨不同结构的利弊，因此不再过多说明。

## 日志
日志是mkDB很重要的组成部分，它记录着mkDB的运行状态，因为这是玩具数据库，所以只记录输入输出的日志。
每次读写时会先写一个数据和，如果开头和结尾的数据和一致，那么说明数据库读写正常

#### undolog
原子性：一个操作不可分割（通常只业务上的一个操作，可能会涉及到多个sql操作）

一致性：事务开始和结束，对于数据来说结果是一致的（从数据层面的原子性）。

隔离性：当前事务不会受到其他事务的影响。

持久性：事务一旦提交，对数据的改变是永久的，即使出现宕机也能恢复。

高并发场景事务带来的问题
脏写：一个事务的操作覆盖了第二个事务的结果。

如：事务A对字段a查询结果为1，然后进行了业务逻辑操作，使a=0，最后将a=0更新到数据库了。但是在更新之前，可能出现事务B已经将a更新成了2，这样就影响了其他事务的更新。

脏读：一个事务查询到了其他事务未提交的数据。

不可重复读：一个事务对一条或多条数据进行了多次查询，发现数据有删除或修改（强调的数据的删除和修改）。

幻读：一个事务对同一批数据进行了多次查询，发现数据有新增（强调数据的新增）。

## 锁
为了简易实现，直接对数据库上锁。话不多说，就是为了支持多线程。虽然慢但是安全。

------------------------------------------

## 数据库相关
(XXX)代表数据库名称
#### CREATE DATABASE XXX
对应接口
* DataBase* DatabaseManager::create(std::string database_name, int mode);
mode表示创建策略，默认值为 0
值为0时表示追加创建，如果有同名数据库则直接连接原有数据库。
值为1时表示覆盖创建，如果有同名数据库则删除原有数据库并新建数据库。
此接口会抛出异常

#### DROP DATABASE XXX
对应接口
* void DatabaseManager::drop(std::string database_name);

#### USE XXX
对应接口
* DataBase* DatabaseManager::connect(std::string database_name);
此接口会抛出异常

## 表相关
(XXX)代表表名称
...代表表相关参数
#### CREATE TABLE XXX(...)
对应接口
* TABLE* TABLE<...>::create(std::string table_name, int mode);
通过模板类实现表
mode表示创建策略，默认值为 0
值为0时表示追加创建，如果有同名表则读取原有表。
值为1时表示覆盖创建，如果有同名表则删除原有表并新建表。

#### DROP TABLE XXX
对应接口
* void TABLE::drop(std::string table_name)

#### INSERT INTO XXX VALUES (...)
对应接口
* int TABLE::insert(...);

#### UPDATE XXX SET ...
对应接口
* int TABLE::update(...);

#### DELETE FROM XXX WHERE ...
对应接口
* int TABLE::remove(...)
此处使用remove的原因是delete作为关键字无法使用。

-----------------------------------------------------------