// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"testing"
	"reflect"
)

type C interface {
	Foo()
}
type D struct {
	E1 string
	E2 string
}
func (d *D) Foo() {
}

func T0() string {
	return ""
}

func T1() (string, error) {
	return "", nil
}

func TestConfigure(t *testing.T) {
	c := New()
	if err := c.Register("T0", T0); err != nil {
		t.Errorf(`Register(T0): %v`, err)
	}
	if err := c.Register("T1", T1); err == nil {
		t.Errorf(`Register(T1) expected an error, got nil`)
	}
	if err := c.Register("T2", "abc"); err == nil {
		t.Errorf(`Register(T2) expected an error, got nil`)
	}
}

func TestConfigureScalar(t *testing.T) {
	type MyBool bool

	tests := []struct {
		json     string
		val      interface{}
		expected interface{}
	}{
		{"true", new(bool), true},
		{"null", new(bool), false},
		{"true", new(MyBool), MyBool(true)},
		{"true", new(interface{}), true},
		{"-100", new(int), int(-100)},
		{"-10", new(int8), int8(-10)},
		{"-1000", new(int16), int16(-1000)},
		{"-100000", new(int64), int64(-100000)},
		{"10", new(uint), uint(10)},
		{"10.1", new(uint), uint(10)},
		{"10.6", new(uint), uint(10)},
		{"11", new(uint8), uint8(11)},
		{"12", new(uint16), uint16(12)},
		{"13", new(uint32), uint32(13)},
		{"14", new(uint64), uint64(14)},
		{"15", new(uintptr), uintptr(15)},
		{"1.2", new(float32), float32(1.2)},
		{"2.3", new(float64), float64(2.3)},
		{`"abc"`, new(string), "abc"},
		{`null`, new(string), ""},
		{`"abc"`, new(interface{}), "abc"},
		{`"abc"`, new([]byte), "abc"},
		{"true", new(*bool), true},
		{"true", new(*MyBool), MyBool(true)},
		{`"abc"`, new(*interface{}), "abc"},
	}
	for _, test := range tests {
		config := New()
		config.LoadJSON([]byte(test.json))
		if err := config.Configure(test.val); err != nil {
			t.Errorf("Configure(%v): %v", test.json, err)
			continue
		}

		// pointer types
		if reflect.TypeOf(test.val).Elem().Kind() == reflect.Ptr {
			v := reflect.ValueOf(test.val).Elem().Elem().Interface()
			if v != test.expected {
				t.Errorf("Configure(%v) = %v, expected %v", test.json, v, test.expected)
			}
			continue
		}

		// byte slice
		v := reflect.ValueOf(test.val).Elem().Interface()
		if reflect.TypeOf(test.val).Elem() == reflect.TypeOf([]byte(nil)) {
			v = string(v.([]byte))
		}

		// non-pointer types
		if v != test.expected {
			t.Errorf("Configure(%v) = %v, expected %v", test.json, v, test.expected)
		}
	}
}

func TestConfigureArray(t *testing.T) {
	type T1 struct {
		A1 string
		A2 int
	}
	tests := []struct {
		kind     string
		json     string
		val      interface{}
		expected interface{}
	}{
		{"[3]int", `[1, 3, 2]`, new([3]int), [3]int{1, 3, 2}},
		{"[3]int", `[1, 3, 2, 4]`, new([3]int), [3]int{1, 3, 2}},
		{"[3]int", `[1, 3]`, new([3]int), [3]int{1, 3, 0}},
		{"[3]int", `null`, new([3]int), [3]int{0, 0, 0}},
		{"[]int", `[10, 30, 20]`, make([]int, 3), []int{10, 30, 20}},
		{"[]int", `[10, 30, 20, 40]`, make([]int, 3), []int{10, 30, 20, 40}},
		{"[]int", `[10, 30]`, make([]int, 3), []int{10, 30}},
		{"[]int", `null`, make([]int, 3), []int{}},
		{"[]int", `[10.1, 30]`, make([]int, 3), []int{10, 30}},
		{"[]interface", `[true, "abc", null, 2.1]`, make([]interface{}, 3), []interface{}{true, "abc", nil, 2.1}},
		{"nil", `[true, false]`, interface{}(nil), [2]bool{true, false}},
		{"[]struct", `[{"A1":"a1", "A2":1}, {"A1":"a2", "A2":2}]`, make([]T1, 3), []T1{{"a1", 1}, {"a2", 2}}},
	}
	for _, test := range tests {
		config := New()
		config.LoadJSON([]byte(test.json))
		var err error
		switch test.kind {
		case "[3]int":
			err = config.Configure(test.val.(*[3]int))
		case "[]int":
			t := test.val.([]int)
			err = config.Configure(&t)
			test.val = t
		case "[]interface":
			t := test.val.([]interface{})
			err = config.Configure(&t)
			test.val = t
		case "[]struct":
			t := test.val.([]T1)
			err = config.Configure(&t)
			test.val = t
		case "nil":
			err = config.Configure(&test.val)
		}

		if err != nil {
			t.Errorf("Configure(%v): %v", test.json, err)
			continue
		}
		v := reflect.ValueOf(test.val)
		for v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		expected := reflect.ValueOf(test.expected)
		if v.Len() != expected.Len() {
			t.Errorf("Configure(%v): got array length %v, expected %v", test.json, v.Len(), expected.Len())
		}
		for i := 0; i < v.Len(); i++ {
			if v.Index(i).Interface() != expected.Index(i).Interface() {
				t.Errorf("Configure(%v): got %v, expected %v", test.json, v.Interface(), test.expected)
				break
			}
		}
	}
}

func TestConfigureMap(t *testing.T) {
	c := New()
	data := []byte(`{
		"A1": true,
		"A2": 100,
		"A3": 1.23,
		"A4": "abc",
		"A5": true,
		"A6": [3, 2],
		"A7": {"a1":1, "a3":3},
		"A8": {"b1":1, "b2":"abc"}
	}`)
	c.LoadJSON(data)
	var obj struct {
		A0 string
		A1 bool
		A2 int
		A3 float64
		A4 string
		A5 *bool
		A6 []int
		A7 map[string]int
		A8 map[string]interface{}
	}
	obj.A0 = "xyz"
	if err := c.Configure(&obj); err != nil {
		t.Errorf("Configure(%v): %v", string(data), err)
		return
	}

	if obj.A0 != "xyz" {
		t.Errorf("obj.A0 = %v, expected %v", obj.A0, "xyz")
	}
	if obj.A1 != true {
		t.Errorf("obj.A1 = %v, expected %v", obj.A1, true)
	}
	if obj.A2 != 100 {
		t.Errorf("obj.A2 = %v, expected %v", obj.A2, 100)
	}
	if obj.A3 != 1.23 {
		t.Errorf("obj.A3 = %v, expected %v", obj.A3, 1.23)
	}
	if obj.A4 != "abc" {
		t.Errorf("obj.A4 = %v, expected %v", obj.A4, "abc")
	}
	if *obj.A5 != true {
		t.Errorf("obj.A5 = %v, expected %v", *obj.A5, true)
	}
	if len(obj.A6) != 2 {
		t.Errorf("len(obj.A6) = %v, expected %v", len(obj.A6), 2)
	} else if obj.A6[0] != 3 || obj.A6[1] != 2 {
		t.Errorf("obj.A6 = %v, expected %v", obj.A6, `[3 2]`)
	}
	if len(obj.A7) != 2 {
		t.Errorf("len(obj.A7) = %v, expected %v", len(obj.A7), 2)
	} else if obj.A7["a1"] != 1 || obj.A7["a3"] != 3 {
		t.Errorf("obj.A7 = %v, expected %v", obj.A7, `map[a1:1 a3:3]`)
	}
	if len(obj.A8) != 2 {
		t.Errorf("len(obj.A8) = %v, expected %v", len(obj.A8), 2)
	} else if obj.A8["b1"] != float64(1) || obj.A8["b2"] != "abc" {
		t.Errorf("obj.A8 = %v, expected %v", obj.A8, `map[b1:1 b2:abc]`)
	}
}

func TestConfigureNested(t *testing.T) {
	c := New()
	data := []byte(`{
		"A1": 100,
		"A2": {
			"B2": "v1",
			"B1": {
				"C1": 300
			}
		},
		"A3": [
			{"B3": true},
			{"B3": false}
		]
	}`)
	c.LoadJSON(data)
	var obj struct {
		A1 int
		A2 *struct {
			B1 struct {
				   C1 int
				   C2 string
			   }
			B2 string
		}
		A3 []struct{
			B3 bool
		}
	}
	if err := c.Configure(&obj); err != nil {
		t.Errorf("Configure(%v): %v", string(data), err)
		return
	}
	if obj.A1 != 100 {
		t.Errorf("obj.A1 = %v, expected %v", obj.A1, 100)
	}
	if obj.A2.B2 != "v1" {
		t.Errorf("obj.A2.B2 = %q, expected %q", obj.A2.B2, "v1")
	}
	if obj.A2.B1.C1 != 300 {
		t.Errorf("obj.A2.B1.C1 = %v, expected %v", obj.A2.B1.C1, 300)
	}
	if obj.A2.B1.C2 != "" {
		t.Errorf("obj.A2.B1.C2 = %q, expected %q", obj.A2.B1.C2, "")
	}
	if len(obj.A3) != 2 {
		t.Errorf("len(obj.A3) = %v, expected %v", len(obj.A3), 2)
	} else if obj.A3[0].B3 != true || obj.A3[1].B3 != false {
		t.Errorf("obj.A3 = %v, expected %v", obj.A3, `[{true} {false}]`)
	}
}

func TestConfigureMapWithType1(t *testing.T) {
	c := New()
	c.Register("C1", func() *D {
		return &D{}
	})
	data := []byte(`{
		"type": "C1",
		"E1": "abc"
	}`)
	c.LoadJSON(data)
	var object C
	if err := c.Configure(&object); err != nil {
		t.Errorf("Configure(object): %v", err)
		return
	}
	if _, ok := object.(*D); !ok {
		t.Errorf("Object not configured")
		return
	}
	if object.(*D).E1 != "abc" {
		t.Errorf("D.E1=%q, expected %q", object.(*D).E1, "abc")
		return
	}

	var object2 C = &D{"xyz", "123"}
	if err := c.Configure(&object2); err != nil {
		t.Errorf("Configure(object2): %v", err)
		return
	}
	if object2.(*D).E1 != "abc" {
		t.Errorf("D.E1=%q, expected %q", object2.(*D).E1, "abc")
	}
	if object2.(*D).E2 != "123" {
		t.Errorf("D.E2=%q, expected %q", object2.(*D).E2, "123")
	}
}

func TestConfigureMapWithType2(t *testing.T) {
	c := New()
	c.Register("C1", func() *D {
		return &D{}
	})
	data := []byte(`[
		{"type": "C1", "E1": "abc"},
		{"type": "C1", "E1": "xyz"}
	]`)
	c.LoadJSON(data)
	var object []C
	if err := c.Configure(&object); err != nil {
		t.Errorf("Configure(object): %v", err)
		return
	}
	if len(object) != 2 {
		t.Errorf("len(array)=%v, expected %v", len(object), 2)
		return
	}
	if object[0].(*D).E1 != "abc" || object[1].(*D).E1 != "xyz" {
		t.Errorf("array=%q, %q, expected %q, %q", object[0].(*D).E1, object[1].(*D).E1, "abc", "xyz")
		return
	}
}

func TestConfigureYamlConfigure(t *testing.T) {
	c := New()
	err := c.Load("testdata/c3.yaml")
	if err != nil {
		t.Error(err)
		return
	}
	var object C = &D{"xyz", "123"}
	if err := c.Configure(&object, "A6"); err != nil {
		t.Errorf(`Configure(object, "A6"): %v`, err)
		return
	}
	if object.(*D).E1 != "xyz" {
		t.Errorf("D.E1=%q, expected %q", object.(*D).E1, "xyz")
	}
	if object.(*D).E2 != "abc" {
		t.Errorf("D.E2=%q, expected %q", object.(*D).E2, "abc")
	}
}
