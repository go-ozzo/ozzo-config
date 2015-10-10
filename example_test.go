// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config_test

import (
	"fmt"
	"github.com/goseal/config"
)

func Example() {
	var app struct {
		Version string
		Params  map[string]interface{}
	}

	// create a new configuration
	c := config.New()

	// load configuration from app.json and app.dev.json files
	if err := c.Load("app.json", "app.dev.json"); err != nil {
		panic(err)
	}

	// retrieve the Version configuration value
	fmt.Println(c.GetString("Version"))

	// configure the app struct
	if err := c.Configure(&app); err != nil {
		panic(err)
	}
	fmt.Println(app.Params["Key"])
}

func ExampleConfig_Get() {
	var data = []byte(`{
		"A": {
			"B1": "v1",
			"B2": {
				"C1": 300
			},
			"B3": true,
			"B4": [100, "abc"]
		}
	}`)

	c := config.New()
	c.LoadJSON(data)

	fmt.Println(c.Get("A.B1"))
	fmt.Println(c.Get("A.B2.C1"))
	fmt.Println(c.Get("A.D", "not found"))
	fmt.Println(c.GetBool("A.B3"))
	fmt.Println(c.GetInt("A.B4.0"))
	// Output:
	// v1
	// 300
	// not found
	// true
	// 100
}

func ExampleConfig_Set() {
	c := config.New()

	c.Set("A.B", 100)
	c.Set("A.C", true)

	fmt.Println(len(c.Get("A").(map[string]interface{})))
	fmt.Println(c.GetInt("A.B"))
	fmt.Println(c.GetBool("A.C"))
	// Output:
	// 2
	// 100
	// true
}

func ExampleConfig_SetData() {
	data1 := map[string]interface{}{
		"A": "abc",
		"B": "xyz",
	}

	data2 := map[string]interface{}{
		"B": "zzz",
		"C": true,
	}

	c := config.New()
	c.SetData(data1, data2)

	fmt.Println(c.Get("A"))
	fmt.Println(c.Get("B"))
	fmt.Println(c.Get("C"))
	// Output:
	// abc
	// zzz
	// true
}

func ExampleConfig_Load() {
	c := config.New()
	if err := c.Load("app.json", "app.dev.json"); err != nil {
		panic(err)
	}
}

func ExampleConfig_LoadJSON() {
	var data1 = []byte(`{"A":true, "B":100, "C":{"D":"xyz"}}`)
	var data2 = []byte(`{"B":200, "C":{"E":"abc"}}`)

	c := config.New()
	if err := c.LoadJSON(data1, data2); err != nil {
		panic(err)
	}

	fmt.Println(c.Get("A"))
	fmt.Println(c.Get("B"))
	fmt.Println(c.Get("C.D"))
	fmt.Println(c.Get("C.E"))
	// Output:
	// true
	// 200
	// xyz
	// abc
}

func ExampleConfig_Register() {
	c := config.New()
	c.Register("MyType", func() string {
		return "my type"
	})
}

func ExampleConfig_Configure() {
	var app struct {
		Version string
		Params  map[string]interface{}
	}

	c := config.New()
	c.LoadJSON([]byte(`{
		"Version": "1.0-alpha",
		"Params": {
			"DataPath": "/data"
		}
	}`))
	c.Configure(&app)
	fmt.Println(app.Version)
	fmt.Println(app.Params["DataPath"])
	// Output:
	// 1.0-alpha
	// /data
}
