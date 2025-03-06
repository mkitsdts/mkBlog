---
title: create blog
created_time: 2023-07-30 17:45:04
category: blog
tags: problem record
author: mkitsdts
---

# Create blog

## Preliminary Work  

### 一、创建项目文件夹(白嫖服务器)
#### 1、创建Github账号  
GitHub有许多好代码，而且可以免费注册账号，不亏，能学到东西简直血赚。但是Github不仅止于此，它为广大用户提供一个可白嫖的服务器，并可用以搭建blog，所以自行到[Github](https://github.com)注册一个账号吧！
#### 2、创建项目
点击新建项目，名称设置为**你的名称（随意）.github.io，切记，一定要XXX.github.io格式
### 二、安装git、nodejs(配置blog必要的环境)  
#### 1、安装git
进入[git官网](https://git-scm.com)（可直接访问，不需要科学环境）  
![git官网](git官网.png)
点击Download，按默认设置安装
安装好后点击鼠标右键，出现Open Git Bash Here，并点击
![git窗口](git窗口.png)
输入git --version
若正常返回不报错，说明git安装成功
#### 2、安装nodejs
进入[nodejs官网](https://nodejs.org/en/download)（可直接访问，不需要科学环境）
![nodejs官网](nodejs官网.png)
点击Download，并按[教程](https://zhuanlan.zhihu.com/p/442215189)配置环境，这里不再过多赘述  
### 三、安装hexo
创建一个文件夹并命名hexo
进入blog再右键鼠标并点击Open Git bash here并输入
``` bash
$ npm install -g hexo-cli
$ hexo init blog
$ npm install hexo -g
$ hexo -v
```
若正常返回说明安装hexo成功
### 四、安装hexo依赖
输入命令
```bash
$ npm install --save hexo-deployer-git
```
若返回
``` bash
$ added 1 package from 1 contributor in ...s
```
说明准备工作完成
## 部署工作
### 1、安装ssh
进入新建的blog文件夹，右键鼠标点击Open Git Bash Here
输入
``` bash
$ ssh
``` 
再输入(生成密钥)
``` bash
$ ssh-keygen -t rsa
```
连续敲击四次回车后会生成两个文件，分别为
秘钥 id_rsa 和公钥 id_rsa.pub
并储存在**C:/Users/ASUS/.ssh**目录
### 2、新建ssh
以文本格式打开文件 id_rsa.pub并复制全部内容
打开Github主页，点击Setting-SSH and GPG keys-New SSH key
将复制的密钥粘贴到key内，title随意，最后点击Add SSH key
### 3、上传文件
打开hexo文件夹，以文本格式进入_config.xml文件，滑至最低端#deployment并将其修改为
`````` bash
deploy:
  type: git
  repository: https://github.com/你的用户名/XXX.github.io.git  #你的仓库地址
  branch: master
``````
***注意：每个冒号后面需要一个空格***
### 4、提交文件
``` bash
$ npm install hexo-deployer-git --save
$ hexo clean
$ hexo g  
$ hexo d
```
完成这一步操作后，在浏览器地址输入**https://xxx.github.io**就可以访问你的博客啦！
至此，个人博客搭建完成，下一步开始修饰blog
---
因为默认主题很简陋，所以可以更换主题或自定义主题，[主题](https://hexo.io/themes/index.html),在这个网站可以下载
# 下载主题
记得替换名称
``` bash
$ npm install hexo-theme-reimu --save
```
# 使用主题
重命名 hexo-theme-reimu 为 reimu
打开配置文件_config.yml
搜索theme，将theme后面的内容更换为名称
最后，打开git bash输入
``` bash
$ hexo clean
$ hexo g -d
```
刷新一下浏览器就好啦
---
用了一段时间后发现，打开网页好卡，得想办法优化一下。。
首先把首页的图片压缩一下，文件太大会影响打开速度。我用的imagine将png转换成webp格式，图片体积普遍来到了100Kb以下。然后打开网页一看，果然快了许多，但还是做不到秒开，这应该是盗用github网页的原因，毕竟国内连接github的速度有目共睹。
于是乎，便走上了优化打开速度的道路了。
