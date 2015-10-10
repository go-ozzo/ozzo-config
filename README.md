# go-ozzo/config - a configuration handling package for Go

[![Build Status](https://travis-ci.org/go-ozzo/config.svg?branch=master)](https://travis-ci.org/go-ozzo/config)
[![GoDoc](https://godoc.org/github.com/go-ozzo/config?status.png)](http://godoc.org/github.com/go-ozzo/config)

go-ozzo/config is a Go package for handling configurations in Go applications. It supports

* reading JSON (with comments), YAML, and TOML configuration files
* merging multiple configurations
* accessing any part of the configuration
* configuring an object using a part of the configuration

## Requirements

Go 1.2 or above.

## Installation

Run the following command to install the package:

```
go get github.com/go-ozzo/config
```

## Getting Started

The following code snippet shows how you can use this package.

```go
package main

import (
    "fmt"
    "github.com/go-ozzo/config"
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
