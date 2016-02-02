// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
	"strings"
)

// ConfigValueError describes a configuration that cannot be used to configure a target value
type ConfigValueError struct {
	Path    string // path to the configuration value
	Message string // the detailed error message
}

// Error returns the error message represented by ConfigValueError
func (e *ConfigValueError) Error() string {
	path := strings.Trim(e.Path, ".")
	return fmt.Sprintf("%q points to an inappropriate configuration value: %v", path, e.Message)
}

// ConfigTargetError describes a target value that cannot be configured
type ConfigTargetError struct {
	Value reflect.Value
}

// Error returns the error message represented by ConfigTargetError
func (e *ConfigTargetError) Error() string {
	if e.Value.Kind() != reflect.Ptr {
		return "Unable to configure a non-pointer"
	}
	if e.Value.IsNil() {
		return "Unable to configure a nil pointer"
	}
	return ""
}

// ProviderError describes a provider that was not appropriate for a type
type ProviderError struct {
	Value reflect.Value
}

// Error returns the error message represented by ProviderError
func (e *ProviderError) Error() string {
	if e.Value.Kind() != reflect.Func {
		return fmt.Sprintf("The provider should be a function, got %v", e.Value.Kind())
	}
	if e.Value.Type().NumOut() != 1 {
		return fmt.Sprintf("The provider should have a single output, got %v", e.Value.Type().NumOut())
	}
	return ""
}

// Register associates a type name with a provider that creates an instance of the type.
// The provider must be a function with a single output.
// Register is mainly needed when calling Configure() to configure an object and create
// new instances of the specified types.
func (c *Config) Register(name string, provider interface{}) error {
	v := reflect.ValueOf(provider)
	if v.Kind() != reflect.Func || v.Type().NumOut() != 1 {
		return &ProviderError{v}
	}
	c.types[name] = v
	return nil
}

// Configure configures the specified value.
//
// You may configured the value with the whole configuration data or a part of it
// by specifying the path to that part.
//
// To configure a struct, the configuration should be a map. A struct field will be assigned
// with a map value whose key is the same as the field name. If a field is also struct,
// it will be recursively configured with the corresponding map configuration.
//
// When configuring an interface, the configuration should be a map with a special "type" key.
// The "type" element specifies the type name registered by Register(). It allows the method
// to create a correct object given a type name.
//
// Note that the value to be configured must be passed in as a pointer.
// You may specify a path to use a particular part of the configuration to configure
// the value. If a path is not specified, the whole configuration will be used.
func (c *Config) Configure(v interface{}, path ...string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			err = r.(error)
		}
	}()

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &ConfigTargetError{rv}
	}

	p := ""
	config := c.data
	if len(path) > 0 {
		d := c.Get(path[0])
		if d == nil {
			return &ConfigPathError{path[0], "no configuration value was found"}
		}
		p = path[0]
		config = reflect.ValueOf(d)
	}

	return c.configure(rv, config, p)
}

// configure configures the value with the configuration.
func (c *Config) configure(v, config reflect.Value, path string) error {
	// get the concrete value, may allocate space if needed
	v = indirect(v)

	if !v.IsValid() {
		return nil
	}
	for config.Kind() == reflect.Interface || config.Kind() == reflect.Ptr {
		config = config.Elem()
	}

	switch config.Kind() {
	case reflect.Array, reflect.Slice:
		return c.configureArray(v, config, path)
	case reflect.Map:
		switch v.Kind() {
		case reflect.Interface:
			return c.configureInterface(v, config, path)
		case reflect.Struct:
			return c.configureStruct(v, config, path)
		case reflect.Map:
			return c.configureMap(v, config, path)
		default:
			return &ConfigValueError{path, "a map cannot be used to configure " + v.Type().String()}
		}
		return c.configureMap(v, config, path)
	default:
		return c.configureScalar(v, config, path)
	}

	return nil
}

func (c *Config) configureArray(v, config reflect.Value, path string) error {
	vkind := v.Kind()

	// nil interface
	if vkind == reflect.Interface && v.NumMethod() == 0 {
		v.Set(config)
		return nil
	}

	if vkind != reflect.Array && vkind != reflect.Slice {
		return &ConfigValueError{path, fmt.Sprintf("%v cannot be used to configure %v", config.Type(), v.Type())}
	}

	n := config.Len()

	// grow slice if it's smaller than the config array
	if vkind == reflect.Slice && v.Cap() < n {
		t := reflect.MakeSlice(v.Type(), n, n)
		reflect.Copy(t, v)
		v.Set(t)
	}

	if n > v.Cap() {
		n = v.Cap()
	}
	for i := 0; i < n; i++ {
		if err := c.configure(v.Index(i), config.Index(i), path+"."+strconv.Itoa(i)); err != nil {
			return err
		}
	}

	if n < v.Len() {
		if vkind == reflect.Array {
			// Array.  Zero the rest.
			z := reflect.Zero(v.Type().Elem())
			for i := n; i < v.Len(); i++ {
				v.Index(i).Set(z)
			}
		} else {
			v.SetLen(n)
		}
	}

	return nil
}

func (c *Config) configureMap(v, config reflect.Value, path string) error {
	// map must have string kind
	t := v.Type()
	if v.IsNil() {
		v.Set(reflect.MakeMap(t))
	}

	for _, k := range config.MapKeys() {
		elemType := v.Type().Elem()
		mapElem := reflect.New(elemType).Elem()
		if err := c.configure(mapElem, mapIndex(config, k), path+"."+k.String()); err != nil {
			return err
		}
		v.SetMapIndex(k.Convert(v.Type().Key()), mapElem)
	}

	return nil
}

// the "type" field name
var typeKey = reflect.ValueOf("type")

func (c *Config) configureStruct(v, config reflect.Value, path string) error {
	for _, k := range config.MapKeys() {
		if k.String() == typeKey.String() {
			continue
		}
		field := v.FieldByName(k.Interface().(string))
		if !field.IsValid() {
			return &ConfigValueError{path, fmt.Sprintf("field %v not found in struct %v", k.String(), v.Type())}
		}
		if !field.CanSet() {
			return &ConfigValueError{path, fmt.Sprintf("field %v cannot be set", k.String())}
		}
		if field.Kind() == reflect.Ptr {
			if field.IsNil() {
				field.Set(reflect.New(field.Type().Elem()))
			}
			field = field.Elem()
		}
		if err := c.configure(field, mapIndex(config, k), path+"."+k.String()); err != nil {
			return err
		}
	}

	return nil
}

func (c *Config) configureInterface(v, config reflect.Value, path string) error {
	// nil interface
	if v.NumMethod() == 0 {
		v.Set(config)
		return nil
	}

	tk := mapIndex(config, typeKey)
	if !tk.IsValid() {
		return &ConfigValueError{path, "missing the type element"}
	}
	if tk.Kind() != reflect.String {
		return &ConfigValueError{path, "type must be a string"}
	}

	builder, ok := c.types[tk.String()]
	if !ok {
		return &ConfigValueError{path, fmt.Sprintf("type %q is unknown", tk.String())}
	}

	object := builder.Call([]reflect.Value{})[0]

	s := indirect(object)
	if !s.Addr().Type().Implements(v.Type()) {
		return &ConfigValueError{path, fmt.Sprintf("%v does not implement %v", s.Type(), v.Type())}
	}
	v.Set(object)

	return c.configureStruct(s, config, path)
}

func (c *Config) configureScalar(v, config reflect.Value, path string) error {
	if !config.IsValid() {
		switch v.Kind() {
		case reflect.Interface, reflect.Ptr, reflect.Map, reflect.Slice:
			v.Set(reflect.Zero(v.Type()))
			// otherwise, ignore null for primitives/string
		}
		return nil
	}

	// nil interface
	if v.Kind() == reflect.Interface && v.NumMethod() == 0 {
		v.Set(config)
		return nil
	}

	if config.Type().ConvertibleTo(v.Type()) {
		v.Set(config.Convert(v.Type()))
		return nil
	}

	return &ConfigValueError{path, fmt.Sprintf("%v cannot be used to configure %v", config.Type(), v.Type())}
}

func indirect(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
		v = v.Addr()
	}
	for {
		// Load value from interface, but only if the result will be usefully addressable.
		if v.Kind() == reflect.Interface && !v.IsNil() {
			e := v.Elem()
			if e.Kind() == reflect.Ptr && !e.IsNil() {
				v = e
				continue
			}
		}

		if v.Kind() != reflect.Ptr {
			break
		}

		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		v = v.Elem()
	}
	return v
}
