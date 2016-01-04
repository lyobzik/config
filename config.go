package config

import (
	"path"
	"errors"
	"io/ioutil"
	"io"
	"os"
	"time"
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
)

type Config interface {
	GetValue(path string) (value interface{}, err error)

	GetString(path string) (value string, err error)
	GetBool(path string) (value bool, err error)
	GetFloat(path string) (value float64, err error)
	GetInt(path string) (value int64, err error)

	GetStrings(path string, delim string) (value []string, err error)
	GetBools(path string, delim string) (value []bool, err error)
	GetFloats(path string, delim string) (value []float64, err error)
	GetInts(path string, delim string) (value []int64, err error)

	GetConfigPart(path string) (config Config, err error)

	LoadValue(path string, value interface{}) (err error)
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
func GetDuration(c Config, path string) (value time.Duration, err error) {
	stringValue, err := c.GetString(path)
	if err != nil {
		return value, err
	}
	return time.ParseDuration(stringValue)
}

func GetTime(c Config, path string) (value time.Time, err error) {
	return GetTimeFormat(c, path, time.RFC3339)
}

func GetTimeFormat(c Config, path string, format string) (value time.Time, err error) {
	stringValue, err := c.GetString(path)
	if err != nil {
		return value, err
	}
	return time.Parse(format, stringValue)
}

func GetDurations(c Config, path string, delim string) (value []time.Duration, err error) {
	stringValues, err := c.GetStrings(path, delim)
	if err != nil {
		return value, err
	}
	resultValue := make([]time.Duration, len(stringValues))
	for i := range stringValues {
		if resultValue[i], err = time.ParseDuration(stringValues[i]); err != nil {
			return value, err
		}
	}
	return resultValue, nil
}

func GetTimes(c Config, path string, delim string) (value []time.Time, err error) {
	return GetTimesFormat(c, path, time.RFC3339, delim)
}

func GetTimesFormat(c Config, path string, format string, delim string) (value time.Time, err error) {
	stringValues, err := c.GetStrings(path, delim)
	if err != nil {
		return value, err
	}
	resultValue := make([]time.Time, len(stringValues))
	for i := range stringValues {
		if resultValue[i], err = time.Parse(stringValues[i], format); err != nil {
			return value, err
		}
	}
	return resultValue, nil
}
