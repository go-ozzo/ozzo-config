# ozzo-config

[![GoDoc](https://godoc.org/github.com/go-ozzo/ozzo-config?status.png)](http://godoc.org/github.com/go-ozzo/ozzo-config)
[![Build Status](https://travis-ci.org/go-ozzo/ozzo-config.svg?branch=master)](https://travis-ci.org/go-ozzo/ozzo-config)
[![Coverage](http://gocover.io/_badge/github.com/go-ozzo/ozzo-config)](http://gocover.io/github.com/go-ozzo/ozzo-config)

ozzo-config это Go-пакет для работы с конфигурациями в приложениях на Go. Он поддерживает

* чтение конфигуационных файлов JSON (с комментариями), YAML, и TOML
* объединение множественных конфигураций
* доступ к любой части конфигурации
* конфигурирование объектов с помощью частей конфигурации

## Требования

Go 1.2 или выше.

## Установка

Выполните следующие команды для установки:

```
go get github.com/go-ozzo/ozzo-config
```

## С чего начать

Следующий фрагмент кода показывает, как можно использовать этот пакет.

```go
package main

import (
    "fmt"
    "github.com/go-ozzo/ozzo-config"
)

func main() {
    // создаем объект конфигурации Config
    c := config.New()

    // загружаем конфигурацию из строки JSON
    c.LoadJSON([]byte(`{
        "Version": "2.0",
        "Author": {
            "Name": "Foo",
            "Email": "bar@example.com"
        }
    }`))

    // получаем значение "Version", возвращает "1.0" если оно не определено в конфигурации
    version := c.GetString("Version", "1.0")

    var author struct {
        Name, Email string
    }
    // заполнить объект author из конфигурации используя "Author"
    c.Configure(&author, "Author")

    fmt.Println(version)
    fmt.Println(author.Name)
    fmt.Println(author.Email)
    // Вывод:
    // 2.0
    // Foo
    // bar@example.com
}
```

## Загрузка конфигурации

Вы можете загрузить конфигурацию тремя способами:

```go
c := config.New()

// загрузить из одного или нескольких файлов JSON, YAML или TOML.
// форматы файлов определяются по их расширениям: .json, .yaml, .yml, .toml
c.Load("app.json", "app.dev.json")

// загрузить из одной или нескольких JSON строк
c.LoadJSON([]byte(`{"Name": "abc"}`), []byte(`{"Age": 30}`))

// загрузить из одной или нескольких переменных
data1 := struct {
    Name string
} { "abc" }
data2 := struct {
    Age int
} { 30 }
c.SetData(data1, data2)
```

При загрузке из нескольких источников, конфигурация будет получена путем рекурсивного слияния их один за другим.

## Доступ к конфигурации

Вы можете получить доступ к любой части конфигурации, используя один из методов `Get`, таких как` Get () `,` GetString () `,` GetInt () `.
Эти методы требуют параметр пути в формате `X.Y.Z` и обозначают место в конфигурации находящееся по адресу `config["X"]["Y"]["Z"]`. 
Если путь не соответствует значению в конфигурации, то возвращается нулевое или дефолтное значение, если оно указано в методе. 
Например,

```go
// Получить "Author.Email". Значение по умолчанию "bar@example.com"
// которое будет получено, если "Author.Email" не будет найдено в конфигурации.
email := c.GetString("Author.Email", "bar@example.com")
```


## Изменение конфигурации

Вы можете изменить любую часть конфигурации с использованием `Set` метод. Например, следующий код
изменяет значение конфигурации `config["Author"]["Email"]`:

```go
c.Set("Author.Email", "bar@example.com")
```

## Конфигурирование объектов

Вы можете использовать конфигурацию, чтобы настроить свойства объекта. Например, кофигурация такой JSON структуры `{"Name": "Foo", "Email": "bar@example.com"}` 
может быть использована для конфигурирования `Name` и `Email` полей структуры.

При настройке nil интерфейса, вы должны указать конкретный тип в конфигурации при момощи элемента `type` карты конфигурации. 
Тип также должен быть зарегистрирован в начале при помощи вызова `Register()` так что он должен знать как создать конкретный instance.
