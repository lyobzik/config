package config

import (
	"encoding/json"
	"math"
)

type jsonConfig struct {
	data interface{}
}

func newJsonConfig(data []byte) (Config, error) {
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

	return c.GrabValue(path, createJsonValueGrabber(creator, grabber))
}

// Get single value.
func (c *jsonConfig) GetString(path string) (value string, err error) {
	return value, c.GrabValue(path, func(data interface{}) error {
		value, err = parseJsonString(data)
		return err
	})
}

func (c *jsonConfig) GetBool(path string) (value bool, err error) {
	return value, c.GrabValue(path, func(data interface{}) error {
		value, err = parseJsonBool(data)
		return err
	})
}

func (c *jsonConfig) GetFloat(path string) (value float64, err error) {
	return value, c.GrabValue(path, func(data interface{}) error {
		value, err = parseJsonFloat(data)
		return err
	})
}

func (c *jsonConfig) GetInt(path string) (value int64, err error) {
	return value, c.GrabValue(path, func(data interface{}) error {
		value, err = parseJsonInt(data)
		return err
	})
}

// Get array of values.
func (c *jsonConfig) GetStrings(path string, delim string) (value []string, err error) {
	return value, c.GrabValues(path, delim,
		func(cap int) { value = make([]string, 0, cap) },
		func(data interface{}) error {
			var parsed string
			if parsed, err = parseJsonString(data); err == nil {
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
			if parsed, err = parseJsonBool(data); err == nil {
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
			if parsed, err = parseJsonFloat(data); err == nil {
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
			if parsed, err = parseJsonInt(data); err == nil {
				value = append(value, parsed)
			}
			return err
		})
}

// Get subconfig.
func (c *jsonConfig) GetConfigPart(path string) (Config, error) {
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
func parseJsonString(data interface{}) (value string, err error) {
	if value, converted := data.(string); converted {
		return value, err
	}
	return value, ErrorIncorrectValueType
}

func parseJsonBool(data interface{}) (value bool, err error) {
	if value, converted := data.(bool); converted {
		return value, err
	}
	return value, ErrorIncorrectValueType
}

func parseJsonFloat(data interface{}) (value float64, err error) {
	if value, converted := data.(float64); converted {
		return value, err
	}
	return value, ErrorIncorrectValueType
}

func parseJsonInt(data interface{}) (value int64, err error) {
	switch dataValue := data.(type) {
	case int:
		return int64(dataValue), nil
	case int64:
		return int64(dataValue), nil
	case float64:
		// Check that value is integer.
		if math.Abs(math.Trunc(dataValue) - dataValue) < math.Nextafter(0, 1) {
			return int64(dataValue), nil
		}
		return value, ErrorIncorrectValueType
	}
	return value, ErrorIncorrectValueType
}

// Grabbing helpers.
func createJsonValueGrabber(creator ValueSliceCreator, grabber ValueGrabber) ValueGrabber {
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
