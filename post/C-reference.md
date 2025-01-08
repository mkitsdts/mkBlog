---
title: C++ reference
date: 2024-12-10 21:45:13
tags: C/C++
author: mkitsdts
---
在函数调用中，永远逃不开参数传递这个问题。对于一个C++程序，参数传递有着很大的优化空间。

# 参数传递方式

## 值传递
```bash
void func(int x, float y, double *z)
{
    // ...
}
```
func函数的参数都属于值传递的方式，书本上又细分成指针传递和值传递，但这其实都是一回事。func函数的传递方式很常见，参数可以是左值也可以是右值。不管是左值还是右值，调用函数的时候都会构造一个临时变量，然后将参数值传递给这个临时变量。

函数参数x,y保存的是值，无论如何修改这两个变量的值，都不影响传递的参数。
参数z是指针，如果函数修改参数z，是不会对参数造成影响的。但参数z存储的是一个内存地址，如果函数修改的是z储存的地址的值，那么这个修改是会生效的。

形象一点表示，世上有值和指针两兄弟，我们通过克隆技术克隆值和指针两兄弟，暴打两兄弟的克隆体对本体不会造成影响。值是个正常人，但指针不是，指针是一个傀儡，他体内储存着主人的信息，同样的，他的克隆体也储存着主人的信息。我们可以通过傀儡找到主人，实现暴打主人的目的。

下面写一个测试用例。
```bash
#include <iostream>
using namespace std;

void func(int x, float y, double* z)
{
    cout << "\n";
    cout << "func函数开始" << endl;
    cout << "x可以为右值" << endl;
    cout << "f的值: " << y << endl;
    cout << "f的地址: " << &y << endl;
    cout << "z储存的地址: " << z << endl;
    cout << "z的地址: " << &z << endl;
    cout << "z储存的地址的值: " << *z << endl;
    cout << "func函数输出结束" << endl;
    cout << "\n";
    y = 0;
    *z = 0;
    z = nullptr;
}

int main()
{
    float f = 1.5f;
    double d = 3;
    double* z = &d;
    cout << "f的值: " << f << endl;
    cout << "f的地址: " << &f << endl;
    cout << "z储存的地址: " << z << endl;
    cout << "z储存的地址的值: " << *z << endl;
    cout << "\n";
    func(1, f, z);
    cout << "f的值: " << f << endl;
    cout << "f的地址: " << &f << endl;
    cout << "z储存的地址: " << z << endl;
    cout << "z储存的地址的值: " << *z << endl;
    return 0;
}
```
通过上面的分析推断，func函数里f和z的地址与main函数不一样。func函数对y和z的修改不会改变main函数定义的值，因为操作的不是同一块内存。但因为z的值是内存地址，修改该内存地址的值可实现修改main函数的值，*z的输出会从3变成0。如我们所料，程序的确如此。

该传递方式，每次调用时会先对参数进行一次拷贝。比方说有一个自定义结构体作为一个参数，这个结构体占用内存比较大，但在拷贝是仍然会将整个结构体拷贝一份，这会造成比较大的性能损耗，用指针可以避免拷贝，指针的内存大小是固定的，几乎没有性能损耗。

## 值引用

引用是C++额外添加的参数传递方式，在C语言中并不存在。很多人可能会疑惑，为什么要设计引用？

指针作为参数传递固然效率很高，但是存在危险性。引用的作用是降低参数传递的危险性，这就是引用存在的原因。引用更像是一个严格受到编译器限制的指针，引用不能运算，也不能脱离实例存在，保证参数传递效率的同时一定程度上提高了安全性。

引用分为左值引用和右值引用，还有一些技巧。

### 左值引用

左值引用是C++98时候的标准。

```bash
#include <iostream>
using namespace std;

struct A
{
    A()
    {
        cout << "构造函数调用" << endl;
    }
    A(const A& a)
    {
        cout << "拷贝调用" << endl;
    }
    ~A()
    {
        cout << "析构函数调用" << endl;
    }
    void operator = (const A& a)
    {
        cout << "赋值" << endl;
    }
    int x;
};

void func(A &a, A aa)
{
    cout << "a的值" << a.x << endl;
    cout << "a的地址" << &a << endl;
    cout << "aa的值" << aa.x << endl;
    cout << "aa的地址" << &aa << endl;
    cout<<endl;
    a.x = 10;
}

int main()
{
    A a;
    a.x = 1;
    cout << "a的值" << a.x << endl;
    cout << "a的地址" << &a << endl;
    func(a,A());        //第一个参数无需再次构造，而第二个参数多调用了一次拷贝构造
    //func(A(),a);      //1为右值，无法通过编译
    cout << "a的值" << a.x << endl;
    cout << "a的地址" << &a << endl;
    return 0;
}
```

左值引用很好的解决了左值在参数传递过程中的问题，但右值还是存在一些问题，下面介绍C++11加入的右值引用。

### 右值引用

右值在参数传递中还是存在很多问题，值传递无法修改原数据且消耗性能，而且右值难以运用指针。

我们可以给上面的程序改成右值引用。右值引用有&&和常量左值引用两种方式。
```bash
#include <iostream>
using namespace std;

struct A
{
    A()
    {
        cout << "构造函数调用" << endl;
    }
    A(const A& a)
    {
        cout << "拷贝调用" << endl;
    }
    ~A()
    {
        cout << "析构函数调用" << endl;
    }
    void operator = (const A& a)
    {
        cout << "赋值" << endl;
    }
    int x;
};

//可以把A &&aa改成 const A& aa
//改为常量左值引用后无法修改aa的值
void func(A &a, A &&aa)
{
    cout << "a的值" << a.x << endl;
    cout << "a的地址" << &a << endl;
    cout << "aa的值" << aa.x << endl;
    cout << "aa的地址" << &aa << endl;
    cout<<endl;
    a.x = 10;
}

int main()
{
    A a;
    a.x = 1;
    cout << "a的值" << a.x << endl;
    cout << "a的地址" << &a << endl;
    func(a,A{});        //第一个参数无需再次构造，而第二个参数多调用了一次拷贝构造
    //func(A{},a);      //1为右值，无法通过编译
    cout << "a的值" << a.x << endl;
    cout << "a的地址" << &a << endl;
    return 0;
}
```

再次运行会发现，通过右值引用避免了临时变量的拷贝构造。那么现在只剩一个问题，难道每次都要写一个左值引用的函数再写一个右值引用的函数吗？有没有一种办法能同时接受左值和右值的引用呢？答案是有的，我们通过模板实现万能引用。

```bash
// 通过引用折叠的规则实现类型推导从而实现万能引用
template <class T>
void func(T && a)
{
    //
}

int main()
{
    func(1);        //可以传右值
    int a = 1;
    func(a);        //可以传左值
    return 0;
}
```

有的时候可能会需要，万能引用并区分左右值。这个时候需要完美转发，使用std::forward<T>(a)即可实现区分左右值。


总结： 基础数据类型之间传递不需要过多考虑性能损耗，根据实际需求看传递值还是地址选择合适的传递方式即可，尽量避免指针的传递。在更大型的结构体中，需要考虑拷贝带来的性能损耗，建议使用后面提到的引用。