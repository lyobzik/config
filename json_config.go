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
	"encoding/json"
	"math"
)

type jsonConfig struct {
	data interface{}
}

func newJSONConfig(data []byte) (Config, error) {
	var config jsonConfig

	if err := json.Unmarshal(data, &config.data); err != nil {
		return nil, err
	}
	return &config, nil
}

// Grabbers.
func (c *jsonConfig) GrabValue(path string, grabber ValueGrabber) (err error) {
	element, err := c.findElement(path)
	if err != nil {
		return err
	}
	return grabber(element)
}

func (c *jsonConfig) GrabValues(path string, delim string,
	creator ValueSliceCreator, grabber ValueGrabber) (err error) {

	return c.GrabValue(path, createJSONValueGrabber(creator, grabber))
}

// Get single value.
func (c *jsonConfig) GetString(path string) (value string, err error) {
	return value, c.GrabValue(path, func(data interface{}) error {
		value, err = parseJSONString(data)
		return err
	})
}

func (c *jsonConfig) GetBool(path string) (value bool, err error) {
	return value, c.GrabValue(path, func(data interface{}) error {
		value, err = parseJSONBool(data)
		return err
	})
}

func (c *jsonConfig) GetFloat(path string) (value float64, err error) {
	return value, c.GrabValue(path, func(data interface{}) error {
		value, err = parseJSONFloat(data)
		return err
	})
}

func (c *jsonConfig) GetInt(path string) (value int64, err error) {
	return value, c.GrabValue(path, func(data interface{}) error {
		value, err = parseJSONInt(data)
		return err
	})
}

// Get array of values.
func (c *jsonConfig) GetStrings(path string, delim string) (value []string, err error) {
	return value, c.GrabValues(path, delim,
		func(cap int) { value = make([]string, 0, cap) },
		func(data interface{}) error {
			var parsed string
			if parsed, err = parseJSONString(data); err == nil {
				value = append(value, parsed)
			}
			return err
		})
}

func (c *jsonConfig) GetBools(path string, delim string) (value []bool, err error) {
	return value, c.GrabValues(path, delim,
		func(cap int) { value = make([]bool, 0, cap) },
		func(data interface{}) error {
			var parsed bool
			if parsed, err = parseJSONBool(data); err == nil {
				value = append(value, parsed)
			}
			return err
		})
}

func (c *jsonConfig) GetFloats(path string, delim string) (value []float64, err error) {
	return value, c.GrabValues(path, delim,
		func(cap int) { value = make([]float64, 0, cap) },
		func(data interface{}) error {
			var parsed float64
			if parsed, err = parseJSONFloat(data); err == nil {
				value = append(value, parsed)
			}
			return err
		})
}

func (c *jsonConfig) GetInts(path string, delim string) (value []int64, err error) {
	return value, c.GrabValues(path, delim,
		func(cap int) { value = make([]int64, 0, cap) },
		func(data interface{}) error {
			var parsed int64
			if parsed, err = parseJSONInt(data); err == nil {
				value = append(value, parsed)
			}
			return err
		})
}

// Get subconfig.
func (c *jsonConfig) GetConfigPart(path string) (Config, error) {
	if len(splitPath(path)) == 0 {
		return c, nil
	}
	element, err := c.findElement(path)
	if err != nil {
		return nil, err
	}
	return &jsonConfig{data: element}, nil
}

// Json helpers.
func (c *jsonConfig) findElement(path string) (interface{}, error) {
	element := c.data
	pathParts := splitPath(path)
	if len(pathParts) == 0 {
		return nil, ErrorNotFound
	}
	for _, pathPart := range pathParts {
		part, converted := element.(map[string]interface{})
		if !converted {
			return nil, ErrorNotFound
		}
		var exist bool
		if element, exist = part[pathPart]; !exist {
			return nil, ErrorNotFound
		}
	}
	return element, nil
}

// Json value parsers.
func parseJSONString(data interface{}) (value string, err error) {
	if value, converted := data.(string); converted {
		return value, err
	}
	return value, ErrorIncorrectValueType
}

func parseJSONBool(data interface{}) (value bool, err error) {
	if value, converted := data.(bool); converted {
		return value, err
	}
	return value, ErrorIncorrectValueType
}

func parseJSONFloat(data interface{}) (value float64, err error) {
	if value, converted := data.(float64); converted {
		return value, err
	}
	return value, ErrorIncorrectValueType
}

func parseJSONInt(data interface{}) (value int64, err error) {
	switch dataValue := data.(type) {
	case int:
		return int64(dataValue), nil
	case int64:
		return int64(dataValue), nil
	case float64:
		// Check that value is integer.
		if math.Abs(math.Trunc(dataValue)-dataValue) < math.Nextafter(0, 1) {
			return int64(dataValue), nil
		}
		return value, ErrorIncorrectValueType
	}
	return value, ErrorIncorrectValueType
}

// Grabbing helpers.
func createJSONValueGrabber(creator ValueSliceCreator, grabber ValueGrabber) ValueGrabber {
	return func(element interface{}) (err error) {
		values, converted := element.([]interface{})
		if !converted {
			return ErrorIncorrectValueType
		}
		creator(len(values))
		for _, value := range values {
			if err = grabber(value); err != nil {
				return err
			}
		}
		return nil
	}
}
