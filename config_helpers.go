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
	if loader := getSingleValueLoader(c, settings, path, value); loader != nil {
		var loadedValue reflect.Value
		if loadedValue, err = loader(); err == nil {
			value.Set(loadedValue)
		}
	} else {
		err = ErrorUnsupportedTypeToLoadValue
	}
	if settings.IgnoreErrors {
		return nil
	}
	return err
}

type valueLoader func() (reflect.Value, error)

func getSingleValueLoader(c Config, settings LoadSettings, path string, value reflect.Value) valueLoader {
	if existCustomLoader(settings, value) {
		if data, err := c.GetString(path); err == nil {
			return func() (reflect.Value, error) {
				err = loadStructValueUsingCustomLoaderByValue(data, settings, value)
				return value, err
			}
		}
	}
	if isLoadable(value) {
		if data, err := c.GetString(path); err == nil {
			return func() (reflect.Value, error) {
				err = loadStructValueUsingLoadableByValue(data, settings, value)
				return value, err
			}
		}
	}
	switch value.Kind() {
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
	case reflect.Slice:
		return getArrayValuesLoader(c, settings, path, value)
	case reflect.Struct:
		return func() (reflect.Value, error) {
			err := loadStructValueByFields(c, settings, path, value)
			return value, err
		}
	}
	return nil
}

func getArrayValuesLoader(c Config, settings LoadSettings, path string, value reflect.Value) valueLoader {
	elementType := value.Type().Elem()
	elementTypeValue := reflect.New(elementType).Elem()
	if existCustomLoader(settings, elementTypeValue) {
		if values, err := c.GetStrings(path, settings.Delim); err == nil {
			loader, _ := settings.Loaders[elementType.String()]
			return createSliceLoader(values, settings, value, loader)
		}
	}
	if isLoadable(elementTypeValue) {
		if values, err := c.GetStrings(path, settings.Delim); err == nil {
			return createSliceLoader(values, settings, value, func(data string) (reflect.Value, error) {
				value := reflect.New(elementType)
				loadableValue, _ := value.Interface().(Loadable)
				err := loadableValue.LoadValueFromConfig(data)
				return value, err
			})
		}
	}
	switch elementType.Kind() {
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

func existCustomLoader(settings LoadSettings, value reflect.Value) bool {
	_, exist := settings.Loaders[value.Type().String()]
	return exist
}

func isLoadable(value reflect.Value) bool {
	loadableType := reflect.TypeOf((*Loadable)(nil)).Elem()
	return value.Type().Implements(loadableType)
}

func createSliceLoader(values []string, settings LoadSettings, value reflect.Value, loader StringValueLoader) valueLoader {
	return func() (reflect.Value, error) {
		var err error
		outputValues := reflect.MakeSlice(value.Type(), 0, len(values))
		for _, data := range values {
			var loadedValue reflect.Value
			loadedValue, err = loader(data)
			if err != nil {
				break;
			}
			outputValues = reflect.Append(outputValues, loadedValue)
		}
		return outputValues, err
	}
}

func loadStructValueUsingCustomLoaderByValue(data string, settings LoadSettings,
	value reflect.Value) error {

	if loader, exist := settings.Loaders[value.Type().String()]; exist {
		loadedValue, err := loader(data)
		if err != nil {
			return err
		}
		value.Set(loadedValue)
	}
	return nil
}

func loadStructValueUsingLoadableByValue(data string, settings LoadSettings,
	value reflect.Value) error {

	if loadableValue, converted := value.Interface().(Loadable); converted {
		return loadableValue.LoadValueFromConfig(data)
	}
	return nil
}

func loadStructValueByFields(c Config, settings LoadSettings, path string,
	value reflect.Value) (err error) {

	for i := 0; i < value.NumField() && (err == nil || settings.IgnoreErrors); i += 1 {
		fieldValue := value.Field(i)
		fieldPath := joinPath(path, getFieldName(value, i))
		err = loadValue(c, settings, fieldPath, fieldValue)
	}
	return err
}

func getFieldName(value reflect.Value, i int) string {
	fieldType := value.Type().Field(i)
	fieldName := fieldType.Tag.Get(TAG_KEY)
	if len(fieldName) != 0 {
		return fieldName
	}
	return fieldType.Name
}
