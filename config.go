package config

import (
	"path"
	"errors"
	"io/ioutil"
	"io"
	"os"
	"time"
	"reflect"
)

const (
	CONF = "conf"
	INI = "ini"
	JSON = "json"
	XML = "xml"
	YAML = "yaml"
	YML = "yml"
)

var (
	ErrorNotFound = errors.New("Not found")
	ErrorIncorrectPath = errors.New("Incorrect path")
	ErrorUnknownConfigType = errors.New("Unknown config type")
	ErrorIncorrectValueType = errors.New("Incorrect value type")
	ErrorUnsupportedFieldType = errors.New("Unsupported field type")
	ErrorIncorrectValueToLoadConfig = errors.New("Inccorect value to load config")
)

type ValueSliceCreator func(length int)
type ValueGrabber func(interface{}) error
type StringValueGrabber func(string) error

type Config interface {
	GetType() string

	GrabValue(path string, grabber ValueGrabber) (err error)
	GrabValues(path string, delim string, creator ValueSliceCreator, grabber ValueGrabber) (err error)

	GetString(path string) (value string, err error)
	GetBool(path string) (value bool, err error)
	GetFloat(path string) (value float64, err error)
	GetInt(path string) (value int64, err error)

	GetStrings(path string, delim string) (value []string, err error)
	GetBools(path string, delim string) (value []bool, err error)
	GetFloats(path string, delim string) (value []float64, err error)
	GetInts(path string, delim string) (value []int64, err error)

	GetConfigPart(path string) (config Config, err error)
}

// Functions to create config object.
func ReadConfig(configPath string) (Config, error) {
	return ReadTypedConfig(configPath, path.Ext(configPath))
}

func ReadTypedConfig(configPath string, configType string) (Config, error) {
	if len(configPath) == 0 {
		return nil, ErrorIncorrectPath
	}
	configFile, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	return ReadConfigFromReader(configFile, configType)
}

func ReadConfigFromReader(configReader io.Reader, configType string) (Config, error) {
	configData, err := ioutil.ReadAll(configReader)
	if err != nil {
		return nil, err
	}
	creator, err := getConfigCreator(configType)
	if err != nil {
		return nil, err
	}
	return creator(configData)
}

// Functions to read values of some specific types.
func GrabStringValue(c Config, path string, grabber StringValueGrabber) (err error) {
	value, err := c.GetString(path)
	if err != nil {
		return err
	}
	return grabber(value)
}

func GrabStringValues(c Config, path string, delim string,
	creator ValueSliceCreator, grabber StringValueGrabber) (err error) {

	values, err := c.GetStrings(path, delim)
	if err != nil {
		return err
	}
	creator(len(values))
	for _, value := range values {
		if err = grabber(value); err != nil {
			return err
		}
	}
	return nil
}

func GetDuration(c Config, path string) (value time.Duration, err error) {
	return value, GrabStringValue(c, path, func(data string) error {
		value, err = time.ParseDuration(data)
		return err
	})
}

func GetTime(c Config, path string) (value time.Time, err error) {
	return GetTimeFormat(c, path, time.RFC3339)
}

func GetTimeFormat(c Config, path string, format string) (value time.Time, err error) {
	return value, GrabStringValue(c, path, func(data string) error {
		value, err = time.Parse(format, data)
		return err
	})
}

func GetDurations(c Config, path string, delim string) (value []time.Duration, err error) {
	return value, GrabStringValues(c, path, delim,
		func(cap int) { value = make([]time.Duration, 0, cap) },
		func(data string) error {
			var parsed time.Duration
			if parsed, err = time.ParseDuration(data); err == nil {
				value = append(value, parsed)
			}
			return err
		})
}

func GetTimes(c Config, path string, delim string) (value []time.Time, err error) {
	return GetTimesFormat(c, path, time.RFC3339, delim)
}

func GetTimesFormat(c Config, path string, format string, delim string) (value []time.Time, err error) {
	return value, GrabStringValues(c, path, delim,
		func(cap int) { value = make([]time.Time, 0, cap) },
		func(data string) error {
			var parsed time.Time
			if parsed, err = time.Parse(data, format); err == nil {
				value = append(value, parsed)
			}
			return err
		})
}

func LoadValue(c Config, path string, value interface{}) (err error) {
	val := reflect.ValueOf(value)
	if val.Kind() != reflect.Ptr || !val.Elem().CanAddr() || !val.Elem().CanSet() {
		return ErrorIncorrectValueToLoadConfig
	}
	return loadValue(c, path, val.Elem())
}
