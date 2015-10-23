// Copyright 2015 Qiang Xue. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package config provides configuration handling for applications.
package config

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/hnakamur/jsonpreprocess"
	"gopkg.in/yaml.v2"
)

// FileTypeError describes the name of a file whose format is not supported.
type FileTypeError string

// Error returns the error message represented by FileTypeError
func (s FileTypeError) Error() string {
	return "File format not supported: " + filepath.Ext(string(s))
}

// ConfigPathError describes a path which cannot be used to set a configuration value.
type ConfigPathError struct {
	Path string
	Message string
}

// Error returns the error message represented by ConfigPathError
func (s *ConfigPathError) Error() string {
	return fmt.Sprintf("%q is not a valid path: %v", s.Path, s.Message)
}

// Config represents a configuration that can be accessed or used to configure objects.
//
// A configuration is a hierarchy of maps and arrays. You may use a path in the dot format
// to access a particular configuration value in the hierarchy. For example, the path
// "Path.To.Xyz" corresponds to the value config["Path"]["To"]["Xyz"].
//
// You can also configure an object with a particular configuration value by calling
// the Configure() method, which sets the object fields with the corresponding configuration value.
//
// Config can be loaded from one or multiple JSON, YAML, or TOML files. Files loaded latter
// will be merged with the earlier ones. You may also directly populate Config with
// the data in memory.
type Config struct {
	data  reflect.Value
	types map[string]reflect.Value
}

// New creates a new Config object.
func New() *Config {
	return &Config{
		types: make(map[string]reflect.Value),
	}
}

// Get retrieves the configuration value corresponding to the specified path.
//
// The path uses a dotted format. A path "Path.To.Xyz" corresponds to the configuration
// value config["Path"]["To"]["Xyz"], provided both config["Path"] and config["Path"]["To"]
// are both valid maps. If not, a default value will be returned. If you do not
// specify a default value, nil will be returned. You may use array keys in the dotted format
// as well to access arrays in the configuration. For example, a path "Path.2.Xyz" corresponds
// to the value config["Path"][2]["Xyz"], if config["Path"] is an array and config["Path"][2]
// is a valid.
//
// If any part of the path corresponds to an invalid value (not a map/array, or is nil),
// the default value will be returned. If you do not specify a default value, nil will be returned.
//
// Note that if you specify a default value, the return value of this method will
// be automatically converted to the same type of the default value.
// If the conversion cannot be conducted, the default value will be returned.
func (c *Config) Get(path string, defaultValue ...interface{}) interface{} {
	// find the actual default value
	var d interface{}
	if len(defaultValue) > 0 {
		d = defaultValue[0]
	}

	// find the config value corresponding to the path
	// if any part of path cannot be located, return the default value
	data := c.data
	parts := strings.Split(path, ".")
	n := len(parts)
	for i := 0; i < n-1; i++ {
		if data = getElement(data, parts[i]); !data.IsValid() {
			return d
		}
	}
	v := getElement(data, parts[n-1])
	if !v.IsValid() {
		return d
	}

	// convert the value to the same type as the default value
	if td := reflect.ValueOf(d); td.IsValid() {
		if v.Type().ConvertibleTo(td.Type()) {
			return v.Convert(td.Type()).Interface()
		}
		// unable to convert: return the default value
		return d
	}
	return v.Interface()
}

// GetString retrieves the string-typed configuration value corresponding to the specified path.
// Please refer to Get for the detailed usage explanation.
func (c *Config) GetString(path string, defaultValue ...string) string {
	var d string
	if len(defaultValue) > 0 {
		d = defaultValue[0]
	}
	return c.Get(path, d).(string)
}

// GetString retrieves the int-typed configuration value corresponding to the specified path.
// Please refer to Get for the detailed usage explanation.
func (c *Config) GetInt(path string, defaultValue ...int) int {
	var d int
	if len(defaultValue) > 0 {
		d = defaultValue[0]
	}
	return c.Get(path, d).(int)
}

// GetString retrieves the int64-typed configuration value corresponding to the specified path.
// Please refer to Get for the detailed usage explanation.
func (c *Config) GetInt64(path string, defaultValue ...int64) int64 {
	var d int64
	if len(defaultValue) > 0 {
		d = defaultValue[0]
	}
	return c.Get(path, d).(int64)
}

// GetString retrieves the float64-typed configuration value corresponding to the specified path.
// Please refer to Get for the detailed usage explanation.
func (c *Config) GetFloat(path string, defaultValue ...float64) float64 {
	var d float64
	if len(defaultValue) > 0 {
		d = defaultValue[0]
	}
	return c.Get(path, d).(float64)
}

// GetString retrieves the bool-typed configuration value corresponding to the specified path.
// Please refer to Get for the detailed usage explanation.
func (c *Config) GetBool(path string, defaultValue ...bool) bool {
	d := false
	if len(defaultValue) > 0 {
		d = defaultValue[0]
	}
	return c.Get(path, d).(bool)
}

// Set sets the configuration value at the specified path.
//
// The path uses a dotted format. A path "Path.To.Xyz" corresponds to the configuration
// value config["Path"]["To"]["Xyz"], while "Path.2.Xyz" corresponds to config["Path"][2]["Xyz"].
// If a value already exists at the specified path, it will be overwritten with the new value.
// If a partial path has no corresponding configuration value, one will be created. For example,
// if the map config["Path"] has no "To" element, a new map config["Path"]["To"] will be created
// so that we can set the value of config["Path"]["To"]["Xyz"].
//
// The method will return an error if it is unable to set the value for various reasons, such as
// the new value cannot be added to the existing array or map.
func (c *Config) Set(path string, value interface{}) error {
	if !c.data.IsValid() {
		c.data = reflect.ValueOf(make(map[string]interface{}))
	}

	data := c.data
	parts := strings.Split(path, ".")
	n := len(parts)
	for i := 0; i < n; i++ {
		switch data.Kind() {
		case reflect.Map, reflect.Slice, reflect.Array:
		default:
			return &ConfigPathError{strings.Join(parts[:i+1], "."), fmt.Sprintf("got %v instead of a map, array, or slice", data.Kind())}
		}

		if i == n-1 {
			if err := setElement(data, parts[i], value); err != nil {
				return &ConfigPathError{path, err.Error()}
			}
			return nil
		}

		e := getElement(data, parts[i])
		if e.IsValid() {
			data = e
			continue
		}

		newMap := make(map[string]interface{})
		if err := setElement(data, parts[i], newMap); err != nil {
			return &ConfigPathError{strings.Join(parts[:i+1], "."), err.Error()}
		}

		data = reflect.ValueOf(newMap)
	}

	return nil
}

// Data returns the complete configuration data.
// Nil will be returned if the configuration has never been loaded before.
func (c *Config) Data() interface{} {
	if c.data.IsValid() {
		return c.data.Interface()
	}
	return nil
}

// SetData sets the configuration data.
//
// If multiple configurations are given, they will be merged sequentially. The following rules are taken
// when merging two configurations C1 and C2:
// A). If either C1 or C2 is not a map, replace C1 with C2;
// B). Otherwise, add all key-value pairs of C2 to C1; If a key of C2 is also found in C1,
// merge the corresponding values in C1 and C2 recursively.
//
// Note that this method will clear any existing configuration data.
func (c *Config) SetData(data ...interface{}) {
	c.data = reflect.Value{}
	for _, d := range data {
		c.data = merge(c.data, reflect.ValueOf(d))
	}
}

// Load loads configuration data from one or multiple files.
//
// If multiple configuration files are given, the corresponding configuration data will be merged
// sequentially according to the rules described in SetData().
//
// Supported configuration file formats include JSON, YAML, and TOML. The file formats
// are determined by the file name extensions (.json, .yaml, .yml, .toml).
// The method will return any file reading or parsing errors.
//
// Note that this method will NOT clear the existing configuration data.
func (c *Config) Load(files ...string) error {
	for _, file := range files {
		var data interface{}
		if err := load(file, &data); err != nil {
			return err
		}
		c.data = merge(c.data, reflect.ValueOf(data))
	}
	return nil
}

// LoadJSON loads new configuration data which are given as JSON strings.
//
// If multiple JSON strings are given, the corresponding configuration data will be merged
// sequentially according to the rules described in SetData().
//
// The method will return any JSON parsing error.
//
// Note that this method will NOT clear the existing configuration data.
func (c *Config) LoadJSON(data ...[]byte) error {
	for _, bytes := range data {
		var err error
		if bytes, err = stripJSONComments(bytes); err != nil {
			return err
		}
		var d interface{}
		if err = json.Unmarshal(bytes, &d); err != nil {
			return err
		}
		c.data = merge(c.data, reflect.ValueOf(d))
	}
	return nil
}

// load reads and parses a JSON, YAML, or TOML file.
func load(file string, data interface{}) error {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	switch strings.ToLower(filepath.Ext(file)) {
	case ".json":
		if bytes, err = stripJSONComments(bytes); err != nil {
			return err
		}
		if err := json.Unmarshal(bytes, data); err != nil {
			return err
		}
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(bytes, data); err != nil {
			return err
		}
	case ".toml":
		if _, err := toml.Decode(string(bytes), data); err != nil {
			return err
		}
	default:
		return FileTypeError(file)
	}

	return nil
}

func merge(v1, v2 reflect.Value) reflect.Value {
	if v1.Kind() != reflect.Map || v2.Kind() != reflect.Map || !v1.IsValid() {
		return v2
	}

	for _, key := range v2.MapKeys() {
		e1 := mapIndex(v1, key)
		e2 := mapIndex(v2, key)
		if e1.Kind() == reflect.Map && e2.Kind() == reflect.Map {
			e2 = merge(e1, e2)
		}
		v1.SetMapIndex(key, e2)
	}

	return v1
}

// mapIndex returns an element value of a map at the specified index.
// If the value is an interface, the underlying value will be returned.
func mapIndex(mp reflect.Value, index reflect.Value) reflect.Value {
	v := mp.MapIndex(index)
	if v.Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v
}

// getElement returns the element value of a map, array, or slice at the specified index.
func getElement(v reflect.Value, p string) reflect.Value {
	switch v.Kind() {
	case reflect.Map:
		return mapIndex(v, reflect.ValueOf(p))
	case reflect.Array, reflect.Slice:
		if i, err := strconv.Atoi(p); err == nil {
			if i >= 0 && i < v.Len() {
				v = v.Index(i)
				for v.Kind() == reflect.Interface {
					v = v.Elem()
				}
				return v
			}
		}
	}
	return reflect.Value{}
}

// setElement ses the element value of a map, array, or slice at the specified index.
func setElement(data reflect.Value, p string, v interface{}) error {
	value := reflect.ValueOf(v)
	switch data.Kind() {
	case reflect.Map:
		key := reflect.ValueOf(p)
		data.SetMapIndex(key, value)
	case reflect.Slice, reflect.Array:
		idx, err := strconv.Atoi(p)
		if err != nil || idx < 0 {
			return fmt.Errorf("%v is not a valid array or slice index", p)
		}
		if data.Kind() == reflect.Slice {
			if idx >= data.Cap() {
				return fmt.Errorf("%v is out of the slice index bound", p)
			}
			data.SetLen(idx + 1)
		} else if idx >= data.Cap() {
			return fmt.Errorf("%v is out of the array index bound", p)
		}
		data.Index(idx).Set(value)
	}
	return nil
}

func stripJSONComments(s []byte) ([]byte, error) {
	var out bytes.Buffer
	reader := bytes.NewBuffer(s)
	if err := jsonpreprocess.WriteCommentTrimmedTo(&out, reader); err != nil {
		return s, err
	}
	return out.Bytes(), nil
}
