package config

import (
	"reflect"
	"strings"
)

const (
	PATH_DELIMITER = "/"
)

// Heplers.
type configCreator func([]byte) (Config, error)

func getConfigCreator(configType string) (configCreator, error) {
	creators := map[string]configCreator{
		CONF: newIniConfig, INI: newIniConfig,
		JSON: newJsonConfig,
		XML: newXmlConfig,
		YAML: newYamlConfig, YML: newYamlConfig}

	if creator, exist := creators[configType]; exist {
		return creator, nil
	}
	return nil, ErrorUnknownConfigType
}

func splitPath(path string) ([]string) {
	path = strings.Trim(path, PATH_DELIMITER)
	if len(path) > 0 {
		return strings.Split(path, PATH_DELIMITER)
	}
	return []string{}
}

func joinPath(pathParts ...string) string {
	path := strings.Join(pathParts, PATH_DELIMITER)
	if strings.HasPrefix(path, PATH_DELIMITER) {
		return path
	}
	return PATH_DELIMITER + path
}

// Load value implementations.
func loadValue(c Config, path string, value reflect.Value) (err error) {
	switch value.Kind() {
	case reflect.Bool:
		err = loadBoolValue(c, path, value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		err = loadIntValue(c, path, value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		err = loadUintValue(c, path, value)
	case reflect.Float32, reflect.Float64:
		err = loadFloatValue(c, path, value)
	case reflect.Slice:
		err = loadSliceValue(c, path, value)
	case reflect.String:
		err = loadStringValue(c, path, value)
	case reflect.Struct:
		err = loadStructValue(c, path, value)
	default:
		return ErrorUnsupportedFieldType
	}
	return err
}

func loadBoolValue(c Config, path string, value reflect.Value) (err error) {
	var result bool
	if result, err = c.GetBool(path); err == nil {
		value.SetBool(result)
	}
	return err
}

func loadIntValue(c Config, path string, value reflect.Value) (err error) {
	var result int64
	if result, err = c.GetInt(path); err == nil {
		value.SetInt(result)
	}
	return err
}

func loadUintValue(c Config, path string, value reflect.Value) (err error) {
	var result int64
	if result, err = c.GetInt(path); err == nil {
		value.SetUint(uint64(result))
	}
	return err
}

func loadFloatValue(c Config, path string, value reflect.Value) (err error) {
	var result float64
	if result, err = c.GetFloat(path); err == nil {
		value.SetFloat(result)
	}
	return err
}

func loadStringValue(c Config, path string, value reflect.Value) (err error) {
	var result string
	if result, err = c.GetString(path); err == nil {
		value.SetString(result)
	}
	return err
}

func loadBoolValues(c Config, path string, value *reflect.Value) (err error) {
	var results []bool
	if results, err = c.GetBools(path, " "); err == nil {
		for _, result := range results {
			*value = reflect.Append(*value, reflect.ValueOf(result))
		}
	}
	return err
}

func loadIntValues(c Config, path string, value *reflect.Value) (err error) {
	var results []int64
	if results, err = c.GetInts(path, " "); err == nil {
		for _, result := range results {
			*value = reflect.Append(*value, reflect.ValueOf(result))
		}
	}
	return err
}

func loadUintValues(c Config, path string, value *reflect.Value) (err error) {
	var results []int64
	if results, err = c.GetInts(path, " "); err == nil {
		for _, result := range results {
			*value = reflect.Append(*value, reflect.ValueOf(result))
		}
	}
	return err
}

func loadFloatValues(c Config, path string, value *reflect.Value) (err error) {
	var results []float64
	if results, err = c.GetFloats(path, " "); err == nil {
		for _, result := range results {
			*value = reflect.Append(*value, reflect.ValueOf(result))
		}
	}
	return err
}

func loadStringValues(c Config, path string, value *reflect.Value) (err error) {
	var results []string
	if results, err = c.GetStrings(path, " "); err == nil {
		for _, result := range results {
			*value = reflect.Append(*value, reflect.ValueOf(result))
		}
	}
	return err
}

func loadSliceValue(c Config, path string, value reflect.Value) (err error) {
	arrayElementType := value.Type().Elem()
	results := reflect.MakeSlice(value.Type(), 0, 1)
	switch arrayElementType.Kind() {
	case reflect.Bool:
		err = loadBoolValues(c, path, &results)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		err = loadIntValues(c, path, &results)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		err = loadUintValues(c, path, &results)
	case reflect.Float32, reflect.Float64:
		err = loadFloatValues(c, path, &results)
	case reflect.String:
		err = loadStringValues(c, path, &results)
	default:
		return ErrorUnsupportedFieldType
	}
	if err == nil {
		value.Set(results)
	}
	return err
}

func loadStructValue(c Config, path string, value reflect.Value) (err error) {
	for i := 0; i < value.NumField() && err == nil; i += 1 {
		fieldValue := value.Field(i)
		fieldType := value.Type().Field(i)
		fieldName := fieldType.Tag.Get(c.GetType())
		if len(fieldName) == 0 {
			fieldName = fieldType.Name
		}
		err = loadValue(c, joinPath(path, fieldName), fieldValue)
	}
	return err
}