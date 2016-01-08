package config

import (
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
func loadValue(c Config, settings LoadSettings, path string, value reflect.Value) (err error) {
	switch value.Kind() {
	case reflect.Bool:
		fallthrough
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fallthrough
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fallthrough
	case reflect.Float32, reflect.Float64:
		fallthrough
	case reflect.String:
		loader := getSingleValueLoader(c, path, value.Kind())
		var loadedValue reflect.Value
		if loadedValue, err = loader(); err == nil{
			value.Set(loadedValue)
		}
	case reflect.Slice:
		err = loadSliceValue(c, settings, path, value)
	case reflect.Struct:
		err = loadStructValue(c, settings, path, value)
	default:
		err = ErrorUnsupportedFieldTypeToLoadValue
	}
	return fixupLoadingError(err, settings)
}

type valueLoader func() (reflect.Value, error)

func getSingleValueLoader(c Config, path string, kind reflect.Kind) valueLoader {
	switch kind {
	case reflect.Bool:
		return func() (reflect.Value, error) {
			value, err := c.GetBool(path)
			return reflect.ValueOf(value), err
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return func() (reflect.Value, error) {
			value, err := c.GetInt(path)
			return reflect.ValueOf(value), err
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return func() (reflect.Value, error) {
			value, err := c.GetInt(path)
			return reflect.ValueOf(uint64(value)), err
		}
	case reflect.Float32, reflect.Float64:
		return func() (reflect.Value, error) {
			value, err := c.GetFloat(path)
			return reflect.ValueOf(value), err
		}
	case reflect.String:
		return func() (reflect.Value, error) {
			value, err := c.GetString(path)
			return reflect.ValueOf(value), err
		}
	}
	return nil
}

func getArrayValuesLoader(c Config, settings LoadSettings, path string, kind reflect.Kind) valueLoader {
	switch kind {
	case reflect.Bool:
		return func() (reflect.Value, error) {
			value, err := c.GetBools(path, settings.Delim)
			return reflect.ValueOf(value), err
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return func() (reflect.Value, error) {
			value, err := c.GetInts(path, settings.Delim)
			return reflect.ValueOf(value), err
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return func() (reflect.Value, error) {
			value, err := c.GetInts(path, settings.Delim)
			convertedValue := make([]uint64, len(value))
			for i := range value {
				convertedValue[i] = uint64(value[i])
			}
			return reflect.ValueOf(convertedValue), err
		}
	case reflect.Float32, reflect.Float64:
		return func() (reflect.Value, error) {
			value, err := c.GetFloats(path, settings.Delim)
			return reflect.ValueOf(value), err
		}
	case reflect.String:
		return func() (reflect.Value, error) {
			value, err := c.GetStrings(path, settings.Delim)
			return reflect.ValueOf(value), err
		}
	}
	return nil
}

func loadSliceValue(c Config, settings LoadSettings, path string, value reflect.Value) (err error) {
	valueKind := value.Type().Elem().Kind()
	switch valueKind {
	case reflect.Bool:
		fallthrough
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		fallthrough
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		fallthrough
	case reflect.Float32, reflect.Float64:
		fallthrough
	case reflect.String:
		loader := getArrayValuesLoader(c, settings, path, valueKind)
		var loadedValue reflect.Value
		if loadedValue, err = loader(); err == nil{
			value.Set(loadedValue)
		}
	default:
		err = ErrorUnsupportedFieldTypeToLoadValue
	}
	return fixupLoadingError(err, settings)
}

func loadStructValue(c Config, settings LoadSettings, path string, value reflect.Value) (err error) {
	// Load struct using custom loader.
	if loader, exist := settings.Loaders[value.Type().String()]; exist {
		if data, err := c.GetString(path); err == nil {
			loadedValue, err := loader(data)
			if err == nil {
				value.Set(loadedValue)
			}
			return fixupLoadingError(err, settings)
		}
	}
	// Load struct using Loadable interface.
	if isLoadable(value) {
		if loadableValue, converted := value.Interface().(Loadable); converted {
			if data, err := c.GetString(path); err == nil {
				err = loadableValue.LoadValueFromConfig(data)
				return fixupLoadingError(err, settings)
			}
		}
	}
	// Load struct by field.
	for i := 0; i < value.NumField() && err == nil; i += 1 {
		fieldValue := value.Field(i)
		fieldPath := joinPath(path, getFieldName(value, i))
		err = loadValue(c, settings, fieldPath, fieldValue)
		err = fixupLoadingError(err, settings)
	}
	return err
}

func isLoadable(value reflect.Value) bool {
	loadableType := reflect.TypeOf((*Loadable)(nil)).Elem()
	return value.Type().Implements(loadableType)
}

func getFieldName(value reflect.Value, i int) string {
	fieldType := value.Type().Field(i)
	fieldName := fieldType.Tag.Get(TAG_KEY)
	if len(fieldName) != 0 {
		return fieldName
	}
	return fieldType.Name
}

func fixupLoadingError(err error, settings LoadSettings) error {
	if settings.IgnoreErrors {
		return nil
	}
	return err
}