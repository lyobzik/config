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
	if existCustomLoader(settings, value) {
		return loadStructValueUsingCustomLoader(c, settings, path, value)
	}
	if isLoadable(value) {
		return loadStructValueUsingLoadable(c, settings, path, value)
	}
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
		err = loadStructValueByFields(c, settings, path, value)
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
	elementType := value.Type().Elem()
	if existCustomLoader(settings, reflect.New(elementType).Elem()) {
		values, err := c.GetStrings(path, settings.Delim)
		if err == nil {
			return err
		}
		loader, _ := settings.Loaders[value.Type().String()]
		outputValues := reflect.MakeSlice(elementType, 0, len(values))
		for _, value := range values {
			loadedValue, err := loader(value)
			err = fixupLoadingError(err, settings)
			if err != nil {
				return err
			}
			outputValues = reflect.Append(outputValues, loadedValue)
		}
		value.Set(outputValues)
		return nil
	}
	if isLoadable(reflect.New(elementType).Elem()) {
		values, err := c.GetStrings(path, settings.Delim)
		if err == nil {
			return err
		}
		outputValues := reflect.MakeSlice(elementType, 0, len(values))
		for _, value := range values {
			outputValue :=  reflect.New(elementType)
			if loadableValue, converted := outputValue.Interface().(Loadable); converted {
				err = loadableValue.LoadValueFromConfig(value)
				err = fixupLoadingError(err, settings)
				if err != nil {
					return err
				}
			}
			outputValues = reflect.Append(outputValues, outputValue)
		}
		value.Set(outputValues)
		return nil
	}
	valueKind := elementType.Kind()
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

func loadStructValueUsingCustomLoader(c Config, settings LoadSettings, path string,
	value reflect.Value) error {

	if loader, exist := settings.Loaders[value.Type().String()]; exist {
		if data, err := c.GetString(path); err == nil {
			loadedValue, err := loader(data)
			if err == nil {
				value.Set(loadedValue)
			}
			return fixupLoadingError(err, settings)
		}
	}
	return nil
}

func loadStructValueUsingLoadable(c Config, settings LoadSettings, path string,
	value reflect.Value) error {

	if loadableValue, converted := value.Interface().(Loadable); converted {
		if data, err := c.GetString(path); err == nil {
			err = loadableValue.LoadValueFromConfig(data)
			return fixupLoadingError(err, settings)
		}
	}
	return nil
}

func loadStructValueByFields(c Config, settings LoadSettings, path string,
	value reflect.Value) (err error) {

	for i := 0; i < value.NumField() && err == nil; i += 1 {
		fieldValue := value.Field(i)
		fieldPath := joinPath(path, getFieldName(value, i))
		err = loadValue(c, settings, fieldPath, fieldValue)
		err = fixupLoadingError(err, settings)
	}
	return err
}

func existCustomLoader(settings LoadSettings, value reflect.Value) bool {
	_, exist := settings.Loaders[value.Type().String()]
	return exist
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