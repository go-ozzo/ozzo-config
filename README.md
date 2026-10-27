# ozzo-config

[![Go Reference](https://pkg.go.dev/badge/github.com/go-ozzo/ozzo-config.svg)](https://pkg.go.dev/github.com/go-ozzo/ozzo-config)
[![CI](https://github.com/go-ozzo/ozzo-config/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/go-ozzo/ozzo-config/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/go-ozzo/ozzo-config/branch/master/graph/badge.svg)](https://codecov.io/gh/go-ozzo/ozzo-config)
[![Go Report](https://goreportcard.com/badge/github.com/go-ozzo/ozzo-config)](https://goreportcard.com/report/github.com/go-ozzo/ozzo-config)

## Other languages

- [Русский](/docs/README-ru.md)
- [简体中文](/docs/README-zh-CN.md) 

## Description

ozzo-config is a Go package for handling configurations in Go applications. It supports

* reading JSON (with comments), YAML, and TOML configuration files
* merging multiple configurations
* accessing any part of the configuration
* configuring an object using a part of the configuration
* adding new configuration file formats

## Requirements

Go 1.21 or above.

## Installation

Run the following command to install the package:

```
go get github.com/go-ozzo/ozzo-config
```

## Getting Started

The following code snippet shows how you can use this package.

```go
package main

import (
    "fmt"
    "github.com/go-ozzo/ozzo-config"
)

func main() {
    // create a Config object
    c := config.New()

    // load configuration from a JSON string
    c.LoadJSON([]byte(`{
        "Version": "2.0",
        "Author": {
            "Name": "Foo",
            "Email": "bar@example.com"
        }
    }`))

    // get the "Version" value, return "1.0" if it doesn't exist in the config
    version := c.GetString("Version", "1.0")

    var author struct {
        Name, Email string
    }
    // populate the author object from the "Author" configuration
    c.Configure(&author, "Author")

    fmt.Println(version)
    fmt.Println(author.Name)
    fmt.Println(author.Email)
    // Output:
    // 2.0
    // Foo
    // bar@example.com
}
```

## Loading Configuration

You can load configuration in three ways:

```go
c := config.New()

// load from one or multiple JSON, YAML, or TOML files.
// file formats are determined by their extensions: .json, .yaml, .yml, .toml
c.Load("app.json", "app.dev.json")

// load from one or multiple JSON strings
c.LoadJSON([]byte(`{"Name": "abc"}`), []byte(`{"Age": 30}`))

// load from one or multiple variables
data1 := struct {
    Name string
} { "abc" }
data2 := struct {
    Age int
} { 30 }
c.SetData(data1, data2)
```

When loading from multiple sources, the configuration will be obtained by merging them one after another recursively.

## Accessing Configuration

You can access any part of the configuration using one of the `Get` methods, such as `Get()`, `GetString()`, `GetInt()`.
These methods require a path parameter which is in the format of `X.Y.Z` and references the configuration value
located at `config["X"]["Y"]["Z"]`. If a path does not correspond to a configuration value, a zero value or an
explicitly specified default value will be returned by the method. For example,

```go
// Retrieves "Author.Email". The default value "bar@example.com"
// should be returned if "Author.Email" is not found in the configuration.
email := c.GetString("Author.Email", "bar@example.com")
```


## Changing Configuration

You can change any part of the configuration using the `Set` method. For example, the following code
changes the configuration value `config["Author"]["Email"]`:

```go
c.Set("Author.Email", "bar@example.com")
```

## Configuring Objects

You can use a configuration to configure the properties of an object. For example, the configuration
corresponding to the JSON structure `{"Name": "Foo", "Email": "bar@example.com"}` can be used to configure
the `Name` and `Email` fields of a struct.

When configuring a nil interface, you have to specify the concrete type in the configuration via a `type` element
in the configuration map. The type should also be registered first by calling `Register()` so that it knows
how to create a concrete instance.

## New Configuration File Formats

ozzo-config supports three configuration file formats out-of-box: JSON (can contain comments), YAML, and TOML.
To support reading new file formats, you should modify the `config.UnmarshalFuncMap` variable by mapping a
new file extension to the corresponding unmarshal function.

## Format Notes

The parsers come from other modules, and `go.mod` names their lowest supported versions. If your application
requires a newer version of one of them, Go builds with that version, and its rules apply.

**TOML** files are parsed by [BurntSushi/toml](https://github.com/BurntSushi/toml). The minimum, v0.3.1, follows
TOML v0.4.0; its v1 releases accept TOML 1.0 syntax such as dotted keys and mixed-type arrays.

**YAML** files are parsed by [gopkg.in/yaml.v2](https://github.com/go-yaml/yaml/tree/v2), with these semantics:

* a mapping is loaded as `map[interface{}]interface{}`;
* `yes`, `no`, `on`, `off`, `y` and `n` are booleans;
* a timestamp such as `2001-12-14` stays a string.

If you prefer YAML 1.2 and the semantics of yaml v3, register its `Unmarshal` function for the YAML extensions
before loading any files:

```go
import yaml "go.yaml.in/yaml/v3"

config.UnmarshalFuncMap[".yaml"] = yaml.Unmarshal
config.UnmarshalFuncMap[".yml"] = yaml.Unmarshal
```

With yaml v3 a mapping with string keys is loaded as `map[string]interface{}`, `yes` is the string `"yes"`, a
timestamp becomes a `time.Time`, and a duplicate key in a mapping is an error.

## Star History

<a href="https://starhistory.io">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.starhistory.io/png?repos=go-ozzo/ozzo-config&style=dark" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.starhistory.io/png?repos=go-ozzo/ozzo-config&style=professional" />
   <img alt="Star History Chart" src="https://api.starhistory.io/png?repos=go-ozzo/ozzo-config" width="800" />
 </picture>
</a>
