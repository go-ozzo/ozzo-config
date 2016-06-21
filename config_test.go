// Copyright 2016 Qiang Xue. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package config

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestGet(t *testing.T) {
	c := New()
	data := []byte(`{
		"a": 100,
		"b": true,
		"c": 1.23,
		"d": {
			"d1": "v1",
			"d3": {
				"d4": 300
			}
		},
		"e": ["abc", true, {"d": 200}]
	}`)
	if err := c.LoadJSON(data); err != nil {
		t.Error(err)
		return
	}

	tests := []struct {
		name     string
		expected interface{}
	}{
		{"a", 100.0},
		{"b", true},
		{"c", 1.23},
		{"d.d1", "v1"},
		{"d.d3.d4", 300.0},
		{"e.0", "abc"},
		{"e.1", true},
		{"e.2.d", 200.0},
		{"f", nil},
		{"f.b", nil},
	}

	for _, test := range tests {
		v := c.Get(test.name)
		if v != test.expected {
			t.Errorf("Get(%q) = %v, expected %v", test.name, v, test.expected)
		}
	}

	// int
	v1 := c.GetInt("a")
	if v1 != 100 {
		t.Errorf("GetInt(%q) = %v, expected %v", "a", v1, 100)
	}
	v1a := c.GetInt("a1", 200)
	if v1a != 200 {
		t.Errorf("GetInt(%q, 200) = %v, expected %v", "a1", v1a, 200)
	}
	// bool
	v2 := c.GetBool("b")
	if !v2 {
		t.Errorf("GetBool(%q) = %v, expected %v", "b", v2, true)
	}
	v2a := c.GetBool("ba", false)
	if v2a {
		t.Errorf("GetBool(%q, false) = %v, expected %v", "ba", v2a, false)
	}
	// string
	v3 := c.GetString("d.d1")
	if v3 != "v1" {
		t.Errorf("GetString(%q) = %q, expected %q", "d.d1", v3, "v1")
	}
	v3a := c.GetString("d.d1a", "v2")
	if v3a != "v2" {
		t.Errorf("GetString(%q, \"v2\") = %q, expected %q", "d.d1a", v3a, "v2")
	}
	// int64
	v4 := c.GetInt64("a")
	if v4 != 100 {
		t.Errorf("GetInt64(%q) = %v, expected %v", "a", v4, 100)
	}
	v4a := c.GetInt64("a1", 200)
	if v4a != 200 {
		t.Errorf("GetInt64(%q, 200) = %v, expected %v", "a1", v4a, 200)
	}
	// float
	v5 := c.GetFloat("c")
	if v5 != 1.23 {
		t.Errorf("GetFloat(%q) = %v, expected %v", "c", v5, 1.23)
	}
	v5a := c.GetFloat("c1", 2.34)
	if v5a != 2.34 {
		t.Errorf("GetFloat(%q, 2.34) = %v, expected %v", "c1", v5a, 2.34)
	}

	// default value
	v6 := c.Get("e", "zero")
	if v6 != "zero" {
		t.Errorf("Get(%q) = %v, expected %v", "e", v6, "zero")
	}

	// using type of the default value
	v7 := c.Get("a", 0).(int)
	if v7 != 100 {
		t.Errorf("Get(%q, 0) = %v, expected %v", "a", v7, 100)
	}

	// unable to convert to the type of the default value
	v8 := c.Get("a", "abc").(string)
	if v8 != "abc" {
		t.Errorf(`Get(%q, "abc") = %q, expected %q`, "a", v8, "abc")
	}
}

func TestSet(t *testing.T) {
	c := New()

	err := c.Set("a.b", 100)
	if err != nil {
		t.Error(err)
	}
	if c.GetInt("a.b") != 100 {
		t.Errorf(`Set("a.b", 100), Get("a.b") = %v, expected %v`, c.GetInt("a.b"), 100)
	}

	err = c.Set("a.b.c", 200)
	if err == nil {
		t.Error(`Set("a.b.c", 200) should return error`)
	}

	err = c.Set("a.c", true)
	if err != nil {
		t.Error(err)
		return
	}
	if c.GetBool("a.c") != true {
		t.Errorf(`Set("a.c", true), Get("a.c") = %v, expected %v`, c.GetBool("a.c"), true)
	}
	if c.GetInt("a.b") != 100 {
		t.Errorf(`Set("a.c", true), Get("a.b") = %v, expected %v`, c.GetInt("a.b"), 100)
	}
}

func TestSetWithError(t *testing.T) {
	c := New()
	c.LoadJSON([]byte(`{
		"A1": 100,
		"A2": [1, 2]
	}`))
	if err := c.Set("A1.A2", 200); err == nil {
		t.Errorf(`Set("A1.A2", 200) should return an error, got nil`)
	}
	if err := c.Set("A2.-1", 200); err == nil {
		t.Errorf(`Set("A2.-1", 200) should return an error, got nil`)
	}
	if err := c.Set("A2.2", 200); err == nil {
		t.Errorf(`Set("A2.-1", 200) should return an error, got nil`)
	}
}

func TestSetData(t *testing.T) {
	d1 := `
		{
		  "A1": "a1",
		  "A2": 2,
		  "A3": true,
		  "A4": 2.13,
		  "A5": null,
		  "A6": {
			"B1": "b1",
			"B2": {
			  "C1": "c1"
			}
		  },
		  "A7": [
			"d1", "d2"
		  ]
		}
	`
	d2 := `
		{
		  "A2": 3,
		  "A5": "a5",
		  "A6": {
			"B2": {
			  "C2": "c2"
			}
		  }
		}
	`
	var c1, c2 interface{}
	json.Unmarshal([]byte(d1), &c1)
	json.Unmarshal([]byte(d2), &c2)

	c := New()
	c.SetData(c1, c2)
	s, _ := json.Marshal(c.Data())
	expected := `{"A1":"a1","A2":3,"A3":true,"A4":2.13,"A5":"a5","A6":{"B1":"b1","B2":{"C1":"c1","C2":"c2"}},"A7":["d1","d2"]}`
	if string(s) != expected {
		t.Errorf(`SetData(c1, c2), result is %v, expected %v`, string(s), expected)
	}

	c.SetData(nil)
	if c.Data() != nil {
		t.Errorf(`SetData(nil), result is %v, expected nil`, c.Data())
	}
}

func TestLoad(t *testing.T) {
	tests := []struct {
		f1, f2, expected string
	}{
		{"testdata/c1.json", "testdata/c2.json", `{"A1":"a1","A2":3,"A3":true,"A4":2.13,"A5":"a5","A6":{"B1":"b1","B2":{"C1":"c1","C2":"c2"}},"A7":["d1","d2"]}`},
		{"testdata/c1.toml", "testdata/c2.toml", `{"A1":"a1","A2":3,"A3":true,"A4":2.13,"A5":"a5","A6":{"B1":"b1","B2":{"C1":"c1","C2":"c2"}},"A7":["d1","d2"]}`},
	}
	for _, test := range tests {
		c := New()
		err := c.Load(test.f1, test.f2)
		if err != nil {
			t.Error(err)
		}
		s, _ := json.Marshal(c.Data())
		if string(s) != test.expected {
			t.Errorf(`Load(%q, %q), result is %v, expected %v`, test.f1, test.f2, string(s), test.expected)
		}
	}
}

func TestLoadYamlFile(t *testing.T) {
	// YAML is tested differently because it loads a hash as map[interface{}]interface{}
	c := New()
	err := c.Load("testdata/c1.yaml", "testdata/c2.yaml")
	if err != nil {
		t.Error(err)
	}
	tests := []struct {
		name     string
		expected interface{}
	}{
		{"A1", "a1"},
		{"A2", 3},
		{"A3", true},
		{"A4", 2.13},
		{"A5", "a5"},
		{"A6.B1", "b1"},
		{"A6.B2.C1", "c1"},
		{"A6.B2.C2", "c2"},
	}
	for _, test := range tests {
		if c.Get(test.name) != test.expected {
			t.Errorf(`Get(%q) = %v, expected %v`, test.name, c.Get(test.name), test.expected)
		}
	}
}

func TestJSONComment(t *testing.T) {
	data := []byte(`{
		/*
		   comments
		 */
		"a": 100,
		"b": true,
		"d": {
			"d1": "v1",
			"d3": {  // comments
				"d4": 300
			}
		},
		"e": ["abc", true, {"d": 200}]
	}`)
	c := New()
	if err := c.LoadJSON(data); err != nil {
		t.Error(err)
		return
	}
}

func TestMerge(t *testing.T) {
	tests := []struct {
		base     string
		update   string
		expected string
	}{
		{
			"null",
			"123",
			"123",
		}, {
			"123",
			"null",
			"null",
		}, {
			"123",
			"124",
			"124",
		}, {
			"123",
			"true",
			"true",
		}, {
			"[1,3,2]",
			"[100,200]",
			"[100,200]",
		}, {
			`{"a":1}`,
			`{"a":true}`,
			`{"a":true}`,
		}, {
			`{"a":1}`,
			`{"b":2}`,
			`{"a":1,"b":2}`,
		}, {
			`{"a":1,"b":2}`,
			`{"b":3,"c":4}`,
			`{"a":1,"b":3,"c":4}`,
		}, {
			`{"a":{"b":1,"c":2},"d":3}`,
			`{"a":{"c":4,"e":5},"d":30}`,
			`{"a":{"b":1,"c":4,"e":5},"d":30}`,
		}, {
			`{"a":[1,2]}`,
			`{"a":[100,200]}`,
			`{"a":[100,200]}`,
		},
	}

	for _, test := range tests {
		var v1, v2 interface{}
		json.Unmarshal([]byte(test.base), &v1)
		json.Unmarshal([]byte(test.update), &v2)
		v := merge(reflect.ValueOf(v1), reflect.ValueOf(v2))
		var s []byte
		if v.IsValid() {
			s, _ = json.Marshal(v.Interface())
		} else {
			s = []byte("null")
		}
		if string(s) != test.expected {
			t.Errorf("merge(%v, %v) = %v, expected %v", test.base, test.update, string(s), test.expected)
		}
	}
}
