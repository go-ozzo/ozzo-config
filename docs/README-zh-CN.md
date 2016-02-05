# ozzo-config

[![GoDoc](https://godoc.org/github.com/go-ozzo/ozzo-config?status.png)](http://godoc.org/github.com/go-ozzo/ozzo-config)
[![Build Status](https://travis-ci.org/go-ozzo/ozzo-config.svg?branch=master)](https://travis-ci.org/go-ozzo/ozzo-config)
[![Coverage](http://gocover.io/_badge/github.com/go-ozzo/ozzo-config)](http://gocover.io/github.com/go-ozzo/ozzo-config)


## 说明

ozzo-config 是用于处理 go 应用里的配置文件的包。它支持以下功能：

* 读取 JSON (可加注释)，YAML，以及 TOML 格式的文件
* 合并多个配置文件
* 访问配置文件的任意片段
* 根据配置文件中的片段，装配一个对象
* 添加新的配置文件格式

## 需求

Go 1.2 或以上。

## 安装

执行以下指令安装此包：

```
go get github.com/go-ozzo/ozzo-config
```

## 准备开始

以下代码片段展示了本包的基本用法

```go
package main

import (
    "fmt"
    "github.com/go-ozzo/ozzo-config"
)

func main() {
    // 创建一个新的 Config 对象
    c := config.New()

    // 从一个 JSON 字符串中读取配置
    c.LoadJSON([]byte(`{
        "Version": "2.0",
        "Author": {
            "Name": "Foo",
            "Email": "bar@example.com"
        }
    }`))

    // 尝试在配置中读取 "Version" 的值，若找不到，则返回默认值 "1.0"
    version := c.GetString("Version", "1.0")

    var author struct {
        Name, Email string
    }
    // 用 "Author" 部分的配置填充 author 对象的属性。
    c.Configure(&author, "Author")

    fmt.Println(version)
    fmt.Println(author.Name)
    fmt.Println(author.Email)
    // 输出：
    // 2.0
    // Foo
    // bar@example.com
}
```

## 加载配置

你可以用以下三种方式读取配置：

```go
c := config.New()

// 从一个或多个 JSON、YAML 或 TOML 文件载入。
// 文件的格式由其文件扩展名确定：`.json`、`.yaml`、`.yml`、`.toml`
c.Load("app.json", "app.dev.json")

// 从一个或多个 JSON 字符串中载入
c.LoadJSON([]byte(`{"Name": "abc"}`), []byte(`{"Age": 30}`))

// 从一个或多个变量中载入
data1 := struct {
    Name string
} { "abc" }
data2 := struct {
    Age int
} { 30 }
c.SetData(data1, data2)
```

当从多个数据源中载入时，最终的配置会由他们依次以递归合并的方式求得。

## 配置的访问操作

这些方法都需要一个格式为 `X.Y.Z` 的路径参数，指向配置中 `config["X"]["Y"]["Z"]` 位置的值。若配置中没有对应此位置的值，则会返回零值，或一个预先显示指定的返回值。举栗：

```go
// 获取 "Author.Email"。默认值为 "bar@example.com"
// 若配置中找不到 "Author.Email" 就会返回默认值。
email := c.GetString("Author.Email", "bar@example.com")
```


## 配置的变更操作

你可以使用 `Set` 方法改变配置对象中的任意组成部分。比如，以下代码用于改变 `config["Author"]["Email"]` 的配置值。

```go
c.Set("Author.Email", "bar@example.com")
```

##  装配对象

可以用配置对象来配置对象^_^。比如，JSON 结构 `{"Name": "Foo", "Email": "bar@example.com"}` 所对应的配置对象，即可用于配置某结构体的 `Name` 和 `Email` 字段。

当配置一个空接口的时候，你必须通过配置映射中的 `type` 元素显式地在配置中指定具体类型。类型也需要先通过 
`Register()` 进行注册，这样它才知道该如何创建一个具体的实例。

## 新的配置文件格式

ozzo-config 原生支持三种配置文件格式：JSON （可以包含注释）、YAML 和 TOML。
如果你想支持读取新的文件格式，你可以修改 `config.UnmarshalFuncMap` 对象，并把一个新的文件扩展名映射给一个用于反编列（译注：
unmarshal，参考[维基百科](https://en.wikipedia.org/wiki/Marshalling_%28computer_science%29)的说法，编列是指把内存中某对象转换为可用于存储或传递的持久化数据格式的行为。反编列即为其反操作，可理解为把配置文本转制为内存对象的过程）的函数。
