---
title: docker
created_time: 2024-10-10
updated_time: 2024-10-10
category: docker
tags: study note
author: mkitsdts
---------------------------------

最近经常看到docker这个词，可是我什么也不知道，所以就去查了一下。一看，这可是好东西，那我的个人博客就用docker来部署吧！
说干就干，直接抛弃github pages，转向云服务器加docker加nginx，重新部署一下博客。
做事之前，总是要先了解清楚docker究竟是一个什么。

# docker介绍
docker作为近几年流行的技术，在此之前已经有虚拟机，物理机等与之相应的概念，三者之间存在一些差异。
物理机，顾名思义，就是一个完整的主机，从硬件到操作系统再到上层软件都是完整的。
而虚拟机则是在物理机运行的操作系统基础上再创建一个新的操作系统。虚拟机在此的作用是将沟通物理硬件与虚拟硬件，然后在虚拟硬件上运行一个镜像，这个镜像只能是完整的操作系统
docker与虚拟机有所不同，docker省略了虚拟硬件这一环节，docker镜像直接与物理机的硬件交互，这使得docker的镜像可以删除内核，只保留环境

如此docker实现了环境隔离并保持了较高的性能，从而得到广泛使用。
docker开发者将docker分成三个部分，分别是容器，镜像和仓库。
## 镜像
镜像就是docker容器的模板，可以简单理解为操作系统减去内核，只剩下桌面，命令，文件管理等部分。镜像静态地保存在仓库里。
## 容器
容器就是提供给用户进行创建的运行环境。容器以镜像为基础，通过本地编写的Dockerfile文件命令，向镜像文件里添加组件，然后创建一个容器，并运行在操作系统上。
## 仓库
仓库存储着各种各样的镜像，用户可以在使用时通过pull命令拉取镜像。一些特殊原因，国内无法直接拉取docker镜像，在拉取前需要检查镜像源是否正常工作。

# docker使用
## 安装
要想使用docker，首先得安装docker，图形界面系统如何安装就不多说了，照着官网文档做就可以。这里演示一下ubuntu22.4.0server版本的安装。
在安装之前先使用自带的apt包管理卸载ubuntu自带的低版本docker
```bash
$ sudo apt-get remove docker docker-engine docker.io containerd runc
```
### 1、更新软件包
不管做什么，上来先更新一下软件包没毛病
```bash
$ sudo apt update
$ sudo apt upgrade
```
经受两板斧过后再开始后面的步骤
### 2、安装docker依赖
Docker在Ubuntu上依赖一些软件包，所以要先安装依赖
```bash
$ sudo apt-get install ca-certificates curl gnupg lsb-release
```
### 3、添加docker密钥
curl是一个负责网络下载的组件，我们通过阿里云添加密钥
```bash
$ curl -fsSL http://mirrors.aliyun.com/docker-ce/linux/ubuntu/gpg | sudo apt-key add -
```
### 4、添加docker软件源
指定docker下载的地址
```bash
$ sudo add-apt-repository "deb [arch=amd64] http://mirrors.aliyun.com/docker-ce/linux/ubuntu $(lsb_release -cs) stable"
```
### 5、下载并安装
```bash
$ sudo apt-get install docker-ce docker-ce-cli containerd.io
```
### 6、运行docker
启动docker
```bash
$ systemctl start docker
```
### 7、安装docker工具
安装好docker工具后需要重启docker服务
```bash
$ sudo apt-get -y install apt-transport-https ca-certificates curl software-properties-common
$ systemctl restart docker
```
到此docker安装完成，我们可以输入sudo docker version检查docker是否正确安装

## 注意事项
docker安装完成后，就可以使用了。
在使用之前，我们需要检查镜像源，默认镜像源Docker Hub在国内无法访问，我们需要修改为国内镜像源。修改镜像源需要通过vim编辑文件，如果没有安装vim请先安装。
如果已经安装vim请忽略
```bash
$ sudo apt install vim
$ vim etc/docker/daemon.json
```
然后会进入到vim的编辑界面，然后加入下面的代码
```bash
"registry-mirrors" : 
[
    "https://dockerproxy.cn",
]
```
完成编辑后按Esc，然后输入:wq!保存并退出。
回到命令行界面重新启动docker，并列出docker信息检查是否成功修改
```bash
$ systemctl restart docker
$ sudo docker info
```
如果看到下面的地址更换完我们之前保存的地址，说明修改成功。可以开始部署了。