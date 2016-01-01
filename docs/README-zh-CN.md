# ozzo-config

[![GoDoc](https://godoc.org/github.com/go-ozzo/ozzo-config?status.png)](http://godoc.org/github.com/go-ozzo/ozzo-config)
[![Build Status](https://travis-ci.org/go-ozzo/ozzo-config.svg?branch=master)](https://travis-ci.org/go-ozzo/ozzo-config)
[![Coverage](http://gocover.io/_badge/github.com/go-ozzo/ozzo-config)](http://gocover.io/github.com/go-ozzo/ozzo-config)

## 其他语言

[简体中文](/docs/README-zh-CN.md) [Русский](/docs/README-ru.md)

## 说明

ozzo-config is a Go package for handling configurations in Go applications. It supports<br>
ozzo-config 是用于处理 go 应用里的配置文件的包。它支持以下功能：

* reading JSON (with comments), YAML, and TOML configuration files
* 读取 JSON (含注释)，YAML，以及 TOML 格式的文件
* merging multiple configurations
* 合并多个配置文件
* accessing any part of the configuration
* 访问配置文件的任意片段
* configuring an object using a part of the configuration
* 根据配置文件中的片段，装配一个对象
* adding new configuration file formats
* 添加新的配置文件格式

## 需求

Go 1.2 或以上。

## 安装

Run the following command to install the package:<br>
执行以下指令安装此包：

```
go get github.com/go-ozzo/ozzo-config
```

## 准备开始

The following code snippet shows how you can use this package.<br>
以下代码片段展示了如何使用这个包

```go
package main

import (
    "fmt"
    "github.com/go-ozzo/ozzo-config"
)

func main() {
    // create a Config object
    // 创建一个新的 Config 对象
    c := config.New()

    // load configuration from a JSON string
    // 读取从一个 JSON 字符串中读取配置
    c.LoadJSON([]byte(`{
        "Version": "2.0",
        "Author": {
            "Name": "Foo",
            "Email": "bar@example.com"
        }
    }`))

    // get the "Version" value, return "1.0" if it doesn't exist in the config
    // 读取 "Version" 的值，若该值在配置中不存在，则返回默认值 "1.0"
    version := c.GetString("Version", "1.0")

    var author struct {
        Name, Email string
    }
    // populate the author object from the "Author" configuration
    // 用 "Author" 配置里的值填充 author 对象。
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

## 读取配置

You can load configuration in three ways:<br>
你可以用以下三种方式读取配置：

```go
c := config.New()

// load from one or multiple JSON, YAML, or TOML files.
// 从一个或多个 JSON、YAML 或 TOML文件载入。
// file formats are determined by their extensions: .json, .yaml, .yml, .toml
// 文件的格式由其文件扩展名确定：.json, .yaml, .yml, .toml
c.Load("app.json", "app.dev.json")

// load from one or multiple JSON strings
// 从一个或多个 JSON 字符串中载入
c.LoadJSON([]byte(`{"Name": "abc"}`), []byte(`{"Age": 30}`))

// load from one or multiple variables
// 从一个或多个变量中载入
data1 := struct {
    Name string
} { "abc" }
data2 := struct {
    Age int
} { 30 }
c.SetData(data1, data2)
```

When loading from multiple sources, the configuration will be obtained by merging them one after another recursively.<br>
当从多个数据源中载入时，他们会一个一个地以递归合并的方式整合进最终的配置。

## 配置の访问

You can access any part of the configuration using one of the `Get` methods, such as `Get()`, `GetString()`, `GetInt()`.<br>
你可以通过 `Get` 方法中的一个来访问配置中的任意组成部分，比如用 `Get()`、`GetString()` 或 `GetInt()`。
These methods require a path parameter which is in the format of `X.Y.Z` and references the configuration value
located at `config["X"]["Y"]["Z"]`. If a path does not correspond to a configuration value, a zero value or an
explicitly specified default value will be returned by the method. For example,<br>
这些方法需要一个格式为 `X.Y.Z` 的路径参数，指向配置中 `config["X"]["Y"]["Z"]` 位置的值。若配置中没有对应此位置的值，则会返回零值，或一个预先显示指定的返回值。举栗：

```go
// Retrieves "Author.Email". The default value "bar@example.com"
// 获取 "Author.Email"。默认值为 "bar@example.com"
// should be returned if "Author.Email" is not found in the configuration.
// 若配置中找不到 "Author.Email" 就会返回默认值。
email := c.GetString("Author.Email", "bar@example.com")
```


## 配置の改变

You can change any part of the configuration using the `Set` method. For example, the following code
changes the configuration value `config["Author"]["Email"]`:<br>
你可以使用 `Set` 方法改变配置文件中的任意组成部分。比如，以下代码就改变了 `config["Author"]["Email"]` 的配置值。

```go
c.Set("Author.Email", "bar@example.com")
```

##  ~~拾捣~~配置对象

You can use a configuration to configure the properties of an object. For example, the configuration
corresponding to the JSON structure `{"Name": "Foo", "Email": "bar@example.com"}` can be used to configure
the `Name` and `Email` fields of a struct.<br>
你可以用配置文件来配置对象的属性值。比如，JSON 结构 `{"Name": "Foo", "Email": "bar@example.com"}` 所对应的配置对象，即可用于配置一个 struct 的 `Name` 和 `Email` 字段。

When configuring a nil interface, you have to specify the concrete type in the configuration via a `type` element
in the configuration map. The type should also be registered first by calling `Register()` so that it knows
how to create a concrete instance.<br>
当配置一个 nil 接口的时候，你必须通过配置映射中的 `type` 元素显式地在配置中指定具体类型。类型也需要先通过 `Register()` 进行注册，这样它才知道该如何创建一个具体的实例。

## 新的配置文件格式

ozzo-config supports three configuration file formats out-of-box: JSON (can contain comments), YAML, and TOML.
To support reading new file formats, you should modify the `config.UnmarshalFuncMap` variable by mapping a
new file extension to the corresponding unmarshal function.<br>
ozzo-config 原生支持三种配置文件格式：JSON （可以包含注释）、YAML 和 TOML。
如果你想支持读取新的文件格式，你可以修改 `config.UnmarshalFuncMap` 对象，并把一个新的文件扩展名映射给一个用于反编列（译注：此术语类似于拆包或打乱原有序列的意思）的函数。
