package config

import (
	"errors"
	"reflect"
	"strings"
)

const (
	PATH_DELIMITER = "/"
	DEFAULT_ARRAY_DELIMITER = " "
	TAG_KEY = "config"
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

func filterPathParts(pathParts []string) []string {
	filteredPathParts := make([]string, 0, len(pathParts))
	for _, pathPart := range pathParts {
		if len(pathPart) > 0 && pathPart != PATH_DELIMITER {
			filteredPathParts = append(filteredPathParts, pathPart)
		}
	}
	return filteredPathParts
}

func splitPath(path string) ([]string) {
	pathParts := strings.Split(path, PATH_DELIMITER)
	return filterPathParts(pathParts)
}

func joinPath(pathParts ...string) string {
	pathParts = filterPathParts(pathParts)
	path := strings.Join(pathParts, PATH_DELIMITER)
	if strings.HasPrefix(path, PATH_DELIMITER) {
		return path
	}
	return PATH_DELIMITER + path
}

// Load value implementations.
func loadValue(c Config, settings LoadSettings, path string, value *reflect.Value) (err error) {
	switch value.Kind() {
	case reflect.Bool:
		err = loadBoolValue(c, settings, path, value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		err = loadIntValue(c, settings, path, value)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		err = loadUintValue(c, settings, path, value)
	case reflect.Float32, reflect.Float64:
		err = loadFloatValue(c, settings, path, value)
	case reflect.Slice:
		err = loadSliceValue(c, settings, path, value)
	case reflect.String:
		err = loadStringValue(c, settings, path, value)
	case reflect.Struct:
		err = loadStructValue(c, settings, path, value)
	default:
		return ErrorUnsupportedFieldTypeToLoadValue
	}
	return err
}

func loadBoolValue(c Config, settings LoadSettings, path string, value *reflect.Value) (err error) {
	var result bool
	if result, err = c.GetBool(path); err == nil {
		value.SetBool(result)
	}
	return err
}

func loadIntValue(c Config, settings LoadSettings, path string, value *reflect.Value) (err error) {
	var result int64
	if result, err = c.GetInt(path); err == nil {
		value.SetInt(result)
	}
	return err
}

func loadUintValue(c Config, settings LoadSettings, path string, value *reflect.Value) (err error) {
	var result int64
	if result, err = c.GetInt(path); err == nil {
		value.SetUint(uint64(result))
	}
	return err
}

func loadFloatValue(c Config, settings LoadSettings, path string, value *reflect.Value) (err error) {
	var result float64
	if result, err = c.GetFloat(path); err == nil {
		value.SetFloat(result)
	}
	return err
}

func loadStringValue(c Config, settings LoadSettings, path string, value *reflect.Value) (err error) {
	var result string
	if result, err = c.GetString(path); err == nil {
		value.SetString(result)
	}
	return err
}

func loadBoolValues(c Config, settings LoadSettings, path string, value *reflect.Value) (err error) {
	var results []bool
	if results, err = c.GetBools(path, " "); err == nil {
		for _, result := range results {
			*value = reflect.Append(*value, reflect.ValueOf(result))
		}
	}
	return err
}

func loadIntValues(c Config, settings LoadSettings, path string, value *reflect.Value) (err error) {
	var results []int64
	if results, err = c.GetInts(path, " "); err == nil {
		for _, result := range results {
			*value = reflect.Append(*value, reflect.ValueOf(result))
		}
	}
	return err
}

func loadUintValues(c Config, settings LoadSettings, path string, value *reflect.Value) (err error) {
	var results []int64
	if results, err = c.GetInts(path, " "); err == nil {
		for _, result := range results {
			*value = reflect.Append(*value, reflect.ValueOf(result))
		}
	}
	return err
}

func loadFloatValues(c Config, settings LoadSettings, path string, value *reflect.Value) (err error) {
	var results []float64
	if results, err = c.GetFloats(path, " "); err == nil {
		for _, result := range results {
			*value = reflect.Append(*value, reflect.ValueOf(result))
		}
	}
	return err
}

func loadStringValues(c Config, settings LoadSettings, path string, value *reflect.Value) (err error) {
	var results []string
	if results, err = c.GetStrings(path, " "); err == nil {
		for _, result := range results {
			*value = reflect.Append(*value, reflect.ValueOf(result))
		}
	}
	return err
}

func loadSliceValue(c Config, settings LoadSettings, path string, value *reflect.Value) (err error) {
	arrayElementType := value.Type().Elem()
	results := reflect.MakeSlice(value.Type(), 0, 1)
	switch arrayElementType.Kind() {
	case reflect.Bool:
		err = loadBoolValues(c, settings, path, &results)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		err = loadIntValues(c, settings, path, &results)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		err = loadUintValues(c, settings, path, &results)
	case reflect.Float32, reflect.Float64:
		err = loadFloatValues(c, settings, path, &results)
	case reflect.String:
		err = loadStringValues(c, settings, path, &results)
	default:
		return ErrorUnsupportedFieldTypeToLoadValue
	}
	if err == nil {
		value.Set(results)
	}
	return err
}

func loadStructValue(c Config, settings LoadSettings, path string, value *reflect.Value) (err error) {
	if loader, exist := settings.Loaders[value.Type().String()]; exist {
		data, err := c.GetString(path)
		if err == nil {
			var loadedValue reflect.Value
			if loadedValue, err = loader(data); err == nil {
				value.Set(loadedValue)
			}
			return err
		}
	}
	if isLoadable(*value) {
		loadValue := (*value).MethodByName("LoadValueFromConfig")
		if loadValue.IsValid() {
			if data, getError := c.GetString(path); getError == nil {
				loadResult := loadValue.Call([]reflect.Value{reflect.ValueOf(data)})
				if len(loadResult) == 1 {
					var converted bool
					if err, converted = loadResult[0].Interface().(error); converted {
						return err
					}
				}
				return errors.New("Cannot load value")
			}
		}
	}
	for i := 0; i < value.NumField() && err == nil; i += 1 {
		fieldValue := value.Field(i)
		fieldType := value.Type().Field(i)
		fieldName := fieldType.Tag.Get(TAG_KEY)
		if len(fieldName) == 0 {
			fieldName = fieldType.Name
		}
		err = loadValue(c, settings, joinPath(path, fieldName), &fieldValue)
		if err != nil && settings.IgnoreErrors {
			err = nil
		}
	}
	return err
}

func isLoadable(value reflect.Value) bool {
	loadableType := reflect.TypeOf((*Loadable)(nil)).Elem()
	return value.Type().Implements(loadableType)
}