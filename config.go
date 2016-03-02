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
	"errors"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"time"
)

// Constants for available config types.
const (
	CONF = "conf"
	INI  = "ini"
	JSON = "json"
	XML  = "xml"
	YAML = "yaml"
	YML  = "yml"
)

// Errors returned from the library.
var (
	ErrorNotFound                       = errors.New("Not found")
	ErrorIncorrectPath                  = errors.New("Incorrect path")
	ErrorUnknownConfigType              = errors.New("Unknown config type")
	ErrorIncorrectValueType             = errors.New("Incorrect value type")
	ErrorUnsupportedTypeToLoadValue     = errors.New("Unsupported field type")
	ErrorIncorrectValueToLoadFromConfig = errors.New("Inccorect value to load from config")
)

// ValueSliceCreator type of function that creates slice for value grabber.
type ValueSliceCreator func(length int)

// ValueGrabber type of function that retrieves value from config.
type ValueGrabber func(interface{}) error

// StringValueGrabber type of function that retrieves string value from config.
type StringValueGrabber func(string) error

// Config represents configuration with convenient access methods.
type Config interface {
	// GrabValue may be used to retrieve single value of complex type.
	GrabValue(path string, grabber ValueGrabber) (err error)
	// GrabValues may be used to retrieve list of values of complex type. Argument 'delim' may
	// be used into method to split list into separate elements.
	GrabValues(path string, delim string, creator ValueSliceCreator, grabber ValueGrabber) (err error)

	// GetString returns string value by specified path.
	GetString(path string) (value string, err error)
	// GetBool returns bool value by specified path.
	GetBool(path string) (value bool, err error)
	// GetFloat returns float value by specified path.
	GetFloat(path string) (value float64, err error)
	// GetInt returns int value by specified path.
	GetInt(path string) (value int64, err error)

	// GetStrings returns list of string by specified path. Argument 'delim' may be used
	// to split list into separate elements.
	GetStrings(path string, delim string) (value []string, err error)
	// GetBools returns list of bool by specified path. Argument 'delim' may be used
	// to split list into separate elements.
	GetBools(path string, delim string) (value []bool, err error)
	// GetFloats returns list of float by specified path. Argument 'delim' may be used
	// to split list into separate elements.
	GetFloats(path string, delim string) (value []float64, err error)
	// GetInts returns list of int by specified path. Argument 'delim' may be used
	// to split list into separate elements.
	GetInts(path string, delim string) (value []int64, err error)

	// GetConfigPart returns as 'Config' config part by specified path.
	GetConfigPart(path string) (config Config, err error)
}

// *** Functions to create config object. ***

// ReadConfig reads and parses config from file. Config type is detected by file extension.
func ReadConfig(configPath string) (Config, error) {
	return ReadTypedConfig(configPath, getConfigType(configPath))
}

// ReadTypedConfig reads and parses config from file of specified type.
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

// ReadConfigFromReader reads and parses config of specified type from reader.
func ReadConfigFromReader(configReader io.Reader, configType string) (Config, error) {
	configData, err := ioutil.ReadAll(configReader)
	if err != nil {
		return nil, err
	}
	return CreateConfig(configData, configType)
}

// CreateConfigFromString creates and parses config of specified type from string.
func CreateConfigFromString(configData string, configType string) (Config, error) {
	return CreateConfig([]byte(configData), configType)
}

// CreateConfig creates and parses config of specified type from byte array.
func CreateConfig(configData []byte, configType string) (Config, error) {
	creator, err := getConfigCreator(configType)
	if err != nil {
		return nil, err
	}
	return creator(configData)
}

// *** Functions to read values of some specific types. ***

// GrabStringValue retrieves value from config using specified grabber.
func GrabStringValue(c Config, path string, grabber StringValueGrabber) (err error) {
	value, err := c.GetString(path)
	if err != nil {
		return err
	}
	return grabber(value)
}

// GrabStringValues retrieves values from config using specified grabber and slice creator.
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

// GetDuration returns duration value from config.
func GetDuration(c Config, path string) (value time.Duration, err error) {
	return value, GrabStringValue(c, path, func(data string) error {
		value, err = time.ParseDuration(data)
		return err
	})
}

// GetTime returns time value from config.
func GetTime(c Config, path string) (value time.Time, err error) {
	return GetTimeFormat(c, path, time.RFC3339)
}

// GetTimeFormat returns time value from config using custom format to parsing.
func GetTimeFormat(c Config, path string, format string) (value time.Time, err error) {
	return value, GrabStringValue(c, path, func(data string) error {
		value, err = time.Parse(format, data)
		return err
	})
}

// GetDurations returns duration values from config. Argument 'delim' may be used
// to split array into separate elements.
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

// GetTimes returns time values from config. Argument 'delim' may be used
// to split array into separate elements.
func GetTimes(c Config, path string, delim string) (value []time.Time, err error) {
	return GetTimesFormat(c, path, time.RFC3339, delim)
}

// GetTimesFormat returns time value from config using custom format to parsing. Argument
// 'delim' may be used to split array into separate elements.
func GetTimesFormat(c Config, path string, format string, delim string) (value []time.Time, err error) {
	return value, GrabStringValues(c, path, delim,
		func(cap int) { value = make([]time.Time, 0, cap) },
		func(data string) error {
			var parsed time.Time
			if parsed, err = time.Parse(format, data); err == nil {
				value = append(value, parsed)
			}
			return err
		})
}

// *** Function to load value of arbitrary type. ***

// Loadable is interface that contains method to load value from config.
//
// LoadValueFromConfig must parse string, change receiver value and return error on fail.
type Loadable interface {
	LoadValueFromConfig(data string) (err error)
}

// StringValueLoader type of function that loads string value from config.
type StringValueLoader func(string) (reflect.Value, error)

// LoadSettings is settings that used to load values from config.
type LoadSettings struct {
	// Delimiter that will be used to split array into separate elements.
	Delim string
	// Flag that specifies whether to ignore errors.
	IgnoreErrors bool
	// Custom loaders.
	Loaders map[string]StringValueLoader
}

// GetDefaultLoadSettings returns settings that may be used to load value from config.
func GetDefaultLoadSettings(ignoreErrors bool) LoadSettings {
	return LoadSettings{Delim: defaultArrayDelimiter,
		IgnoreErrors: ignoreErrors,
		Loaders:      defaultLoaders}
}

var (
	defaultLoaders = map[string]StringValueLoader{
		"time.Time": func(data string) (reflect.Value, error) {
			value, err := time.Parse(time.RFC3339, data)
			if err == nil {
				return reflect.ValueOf(value), nil
			}
			return reflect.ValueOf(nil), err
		},
		"time.Duration": func(data string) (reflect.Value, error) {
			value, err := time.ParseDuration(data)
			if err == nil {
				return reflect.ValueOf(value), nil
			}
			return reflect.ValueOf(nil), err
		}}
)

// LoadValue loads value from config to specified variable. Argument 'value' must be pointer.
// Function can load simple types (bool, int, uint, float, string), arrays of simple types and
// structure. Structure can be loaded field by field, in this case for path construction in
// first place used tag 'config', in second--name of field. Also structure can be loaded using
// custom loader (of type 'StringValueLoader') or 'Loadable' interface.
func LoadValue(c Config, path string, value interface{}) (err error) {
	return parametrizedLoadValue(c, false, path, value)
}

// LoadValueIgnoringErrors loads value from config to variable ignoring some errors (parsing
// errors, absent value, and etc). Argument 'value' must be pointer. Function can load simple
// types (bool, int, uint, float, string), arrays of simple types and structure. Structure can
// be loaded field by field, in this case for path construction in first place used tag 'config',
// in second--name of field. Also structure can be loaded using custom loader (of type
// 'StringValueLoader') or 'Loadable' interface.
func LoadValueIgnoringErrors(c Config, path string, value interface{}) (err error) {
	return parametrizedLoadValue(c, true, path, value)
}

func parametrizedLoadValue(c Config, ignoreErrors bool, path string,
	value interface{}) (err error) {

	return TunedLoadValue(c, GetDefaultLoadSettings(ignoreErrors), path, value)
}

// TunedLoadValue loads value from config to variable using specified settings. Argument 'value'
// must be pointer. Function can load simple types (bool, int, uint, float, string), arrays of
// simple types and structure. Structure can  be loaded field by field, in this case for path
// construction in first place used tag 'config', in second--name of field. Also structure can be
// loaded using custom loader (of type 'StringValueLoader') or 'Loadable' interface.
func TunedLoadValue(c Config, settings LoadSettings, path string, value interface{}) (err error) {
	val := reflect.ValueOf(value)
	if val.Kind() != reflect.Ptr || !val.Elem().CanAddr() || !val.Elem().CanSet() {
		return ErrorIncorrectValueToLoadFromConfig
	}
	return loadValue(c, settings, path, val.Elem())
}
