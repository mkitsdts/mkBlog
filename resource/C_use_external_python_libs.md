---
title: C++ use external python libs
created_time: 2024-10-17
ç: 2024-10-17
category: C/C++
tags: problem record
author: mkitsdts
---

# 前言
使用Qt开发dnd游戏时需要用到大模型，所以需要C++与python（3.8）联动（nodejs也可以，只是本人这次使用的是python）。在联动过程中遇到一些问题，故写下记录。

# 操作过程
首先解决环境问题，可以选择调用系统环境变量里的python，也可以把python打包进项目里。使用环境变量应该不会出现下面提到的问题，可以直接跳转代码部分，只要保证用户拥有python环境，如果担心用户没有安装python，那就把python环境集成进项目里。

## 打包python环境
首先先从[python官网](https://www.python.org/downloads/)下载安装包,并根据提示安装。然后新建一个文件夹存储python环境，再进入python的安装位置，默认是
C:\Users\"你的用户名XXXX"\AppData\Local\Programs\Python\PythonXXX
把里面的文件（Docs,Script,还有python.exe这些文件不需要）都复制到刚刚新建的文件夹里，然后把第一级目录里的python3.dll和另一个python3XX.dll文件复制到libs里面，再复制一个python3XX.dll（XX代表版本）文件并命名为python3XX_d.dll文件，如果要调用的python文件里还包含有外部库，也要复制到与可执行文件同一级目录，否则无法调用。

## Qt调用python代码
解决了环境问题之后，开始撸码。
```bash
    Py_SetPythonHome((const wchar_t *)(L"./include/PythonXXX/"));                   //此处填写python相对可执行文件的路径
    Py_Initialize();                                                                //对python初始化
    PyRun_SimpleString("import sys");                                               //在Python解释器中执行"import sys"命令
    PyRun_SimpleString("sys.path.append('./')");                                    //将当前目录（.）添加到sys.path列表的末尾
    PyObject* pModule = PyImport_ImportModule("XXX");                               //打开调用py文件，参数填写python脚本文件名
    PyObject* pFunc = PyObject_GetAttrString(pModule, "XXX");                       //获取py文件中的函数
    PyObject* pPara = PyTuple_New(X);                                               //填写X参数个数,Tuple表示Python元组对象的结构体
    PyTuple_SetItem(pPara, 0, ...);                                                 //将参数添加进元组对象,对应参数为（元组对象，添加位置，添加元素）
    PyObject* pValue = PyObject_CallObject(pFunc, pPara);                           //调用函数并获取返回结果
    Py_Finalize();                                                                  //结束python调用
```
这一段代码可以实现python的调用

# 报错
在实际调试中可能会遇到许多错误，在这里记录一下目前遇到的错误以及解决方案
## 1、error：error: expected unqualified-id before ';' token
这是因为Python.h里面定义的slots和Qt的slots冲突了，解决方法是进入python文件夹，打开include目录下的object.h文件，并找到下面这个结构体的定义，宏定义改为Q_SLOTS就可以解决
```bash
typedef struct
{
    const char* name;
    int basicsize;
    int itemsize;
    unsigned int flags;
#undef slots            //添加
    PyType_Slot *slots;
#define slots Q_SLOTS
} PyType_Spec;
```
## 2、PyImport_ImportModule("XXX")返回值为nullptr
这个有多种原因
### 第一个原因
你编写的py文件没有放在相应的目录里
### 第二个原因
导入了外部包，而外部包没有包含在目录里

## 3、error：modulenotfounderror: no module named 'encodings'
这个是因为没有设置python路径，要加上第一行Py_SetPythonHome就没问题了

## 4、提示你找不到 pythonXX_d.lib
把libs文件夹下的pythonXX.lib文件，复制到.exe文件同目录下并重命名为pythonXX_d.lib就行（XX为版本号）