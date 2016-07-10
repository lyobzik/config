//Copyright 2016 lyobzik
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package config

import (
	"path"
	"reflect"
	"strings"
)

const (
	// PathDelimiter default path delimiter.
	pathDelimiter = "/"
	// DefaultArrayDelimiter default array delimiter.
	defaultArrayDelimiter = " "
	// TagKey tag name of structure fields.
	tagKey = "config"
)

// Heplers.
type configCreator func([]byte) (Config, error)

func getConfigCreator(configType string) (configCreator, error) {
	creators := map[string]configCreator{
		CONF: newINIConfig, INI: newINIConfig,
		JSON: newJSONConfig,
		XML:  newXMLConfig,
		YAML: newYAMLConfig, YML: newYAMLConfig}

	if creator, exist := creators[configType]; exist {
		return creator, nil
	}
	return nil, ErrorUnknownConfigType
}

func filterPathParts(pathParts []string) []string {
	filteredPathParts := make([]string, 0, len(pathParts))
	for _, pathPart := range pathParts {
		if len(pathPart) > 0 && pathPart != pathDelimiter {
			filteredPathParts = append(filteredPathParts, pathPart)
		}
	}
	return filteredPathParts
}

func splitPath(path string) []string {
	pathParts := strings.Split(path, pathDelimiter)
	return filterPathParts(pathParts)
}

func joinPath(pathParts ...string) string {
	pathParts = filterPathParts(pathParts)
	path := strings.Join(pathParts, pathDelimiter)
	if strings.HasPrefix(path, pathDelimiter) {
		return path
	}
	return pathDelimiter + path
}

func getConfigType(configPath string) string {
	extension := path.Ext(configPath)
	if len(extension) != 0 {
		extension = extension[1:]
	}
	return extension
}

// Load value implementations.
func loadValue(c Config, settings LoadSettings, path string, value reflect.Value) (err error) {
	var loadedValue reflect.Value
	if loadedValue, err = loadSingleValue(c, settings, path, value); err == nil {
		value.Set(loadedValue)
	}
	if settings.IgnoreErrors {
		return nil
	}
	return err
}

func loadSingleValue(c Config, settings LoadSettings, path string, value reflect.Value) (reflect.Value, error) {
	if loader := getCustomLoader(c, settings, value.Type()); loader != nil {
		if data, err := c.GetString(path); err == nil {
			return loader(data, value)
		}
	}
	switch value.Kind() {
	case reflect.Bool:
		value, err := c.GetBool(path)
		return reflect.ValueOf(value), err
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		resultValue, err := c.GetInt(path)
		return reflect.ValueOf(resultValue).Convert(value.Type()), err
	case reflect.Float32, reflect.Float64:
		resultValue, err := c.GetFloat(path)
		return reflect.ValueOf(resultValue).Convert(value.Type()), err
	case reflect.String:
		value, err := c.GetString(path)
		return reflect.ValueOf(value), err
	case reflect.Slice:
		return loadSliceValue(c, settings, path, value)
	case reflect.Struct:
		return loadStructValueByFields(c, settings, path, value)
	}
	return reflect.ValueOf(nil), ErrorUnsupportedTypeToLoadValue
}

func loadSliceValue(c Config, settings LoadSettings, path string, value reflect.Value) (reflect.Value, error) {
	elementType := value.Type().Elem()
	if loader := getCustomLoader(c, settings, elementType); loader != nil {
		if values, err := c.GetStrings(path, settings.Delim); err == nil {
			return loadSlice(values, settings, value, loader)
		}
	}
	switch elementType.Kind() {
	case reflect.Bool:
		value, err := c.GetBools(path, settings.Delim)
		return reflect.ValueOf(value), err
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, err := c.GetInts(path, settings.Delim)
		return copySlice(err, reflect.ValueOf(value), elementType), err
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		value, err := c.GetInts(path, settings.Delim)
		return copySlice(err, reflect.ValueOf(value), elementType), err
	case reflect.Float32:
		value, err := c.GetFloats(path, settings.Delim)
		return copySlice(err, reflect.ValueOf(value), elementType), err
	case reflect.Float64:
		value, err := c.GetFloats(path, settings.Delim)
		return reflect.ValueOf(value), err
	case reflect.String:
		value, err := c.GetStrings(path, settings.Delim)
		return reflect.ValueOf(value), err
	}
	return reflect.ValueOf(nil), ErrorUnsupportedTypeToLoadValue
}

func loadStructValueByFields(c Config, settings LoadSettings, path string,
	value reflect.Value) (result reflect.Value, err error) {

	for i := 0; i < value.NumField() && (err == nil || settings.IgnoreErrors); i++ {
		fieldValue := value.Field(i)
		fieldPath := joinPath(path, getFieldName(value, i))
		err = loadValue(c, settings, fieldPath, fieldValue)
	}
	return value, err
}

type valueLoader func(string, reflect.Value) (reflect.Value, error)

func getCustomLoader(c Config, settings LoadSettings, valueType reflect.Type) valueLoader {
	if loader, exist := settings.Loaders[valueType.String()]; exist {
		return func(data string, value reflect.Value) (reflect.Value, error) {
			loadedValue, err := loader(data)
			if err == nil {
				value.Set(loadedValue)
			}
			return value, err
		}
	} else if isLoadable(valueType) {
		return func(data string, value reflect.Value) (reflect.Value, error) {
			loadableValue, _ := value.Addr().Interface().(Loadable)
			err := loadableValue.LoadValueFromConfig(data)
			return value, err
		}
	}
	return nil
}

func isLoadable(valueType reflect.Type) bool {
	loadableType := reflect.TypeOf((*Loadable)(nil)).Elem()
	return reflect.PtrTo(valueType).Implements(loadableType)
}

func loadSlice(values []string, settings LoadSettings, value reflect.Value,
	loader valueLoader) (reflect.Value, error) {

	var err error
	outputValues := reflect.MakeSlice(value.Type(), len(values), len(values))
	for i, data := range values {
		_, err = loader(data, outputValues.Index(i))
		if err != nil {
			break
		}
	}
	return outputValues, err
}

func copySlice(err error, srcSlice reflect.Value, dstSliceElemType reflect.Type) reflect.Value {
	var dstSlice reflect.Value
	if err == nil {
		dstSlice = reflect.MakeSlice(reflect.SliceOf(dstSliceElemType),
			srcSlice.Len(), srcSlice.Len())
		for i := 0; i < srcSlice.Len(); i += 1 {
			dstSlice.Index(i).Set(srcSlice.Index(i).Convert(dstSliceElemType))
		}
	}
	return dstSlice
}

func getFieldName(value reflect.Value, i int) string {
	fieldType := value.Type().Field(i)
	fieldName := fieldType.Tag.Get(tagKey)
	if len(fieldName) != 0 {
		return fieldName
	}
	return fieldType.Name
}
