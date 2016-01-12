package config

import (
	"math"

	yaml "gopkg.in/yaml.v2"
)

type yamlConfig struct {
	data interface{}
}

func newYamlConfig(data []byte) (Config, error) {
	var config yamlConfig

	if err := yaml.Unmarshal(data, &config.data); err != nil {
		return nil, err
	}
	return &config, nil
}

// Grabbers.
func (c *yamlConfig) GrabValue(path string, grabber ValueGrabber) (err error) {
	if element, err := c.findElement(path); err == nil {
		return grabber(element)
	} else {
		return err
	}
}

func (c *yamlConfig) GrabValues(path string, delim string,
	creator ValueSliceCreator, grabber ValueGrabber) (err error) {

	element, err := c.findElement(path)
	if err != nil {
		return err
	}
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

// Get single value.
func (c *yamlConfig) GetString(path string) (value string, err error) {
	return value, c.GrabValue(path, func(data interface{}) error {
		value, err = parseYamlString(data)
		return err
	})
}

func (c *yamlConfig) GetBool(path string) (value bool, err error) {
	return value, c.GrabValue(path, func(data interface{}) error {
		value, err = parseYamlBool(data)
		return err
	})
}

func (c *yamlConfig) GetFloat(path string) (value float64, err error) {
	return value, c.GrabValue(path, func(data interface{}) error {
		value, err = parseYamlFloat(data)
		return err
	})
}

func (c *yamlConfig) GetInt(path string) (value int64, err error) {
	return value, c.GrabValue(path, func(data interface{}) error {
		value, err = parseYamlInt(data)
		return err
	})
}

// Get array of values.
func (c *yamlConfig) GetStrings(path string, delim string) (value []string, err error) {
	return value, c.GrabValues(path, delim,
		func(cap int) { value = make([]string, 0, cap) },
		func(data interface{}) error {
			var parsed string
			if parsed, err = parseYamlString(data); err == nil {
				value = append(value, parsed)
			}
			return err
		})
}

func (c *yamlConfig) GetBools(path string, delim string) (value []bool, err error) {
	return value, c.GrabValues(path, delim,
		func(cap int) { value = make([]bool, 0, cap) },
		func(data interface{}) error {
			var parsed bool
			if parsed, err = parseYamlBool(data); err == nil {
				value = append(value, parsed)
			}
			return err
		})
}

func (c *yamlConfig) GetFloats(path string, delim string) (value []float64, err error) {
	return value, c.GrabValues(path, delim,
		func(cap int) { value = make([]float64, 0, cap) },
		func(data interface{}) error {
			var parsed float64
			if parsed, err = parseYamlFloat(data); err == nil {
				value = append(value, parsed)
			}
			return err
		})
}

func (c *yamlConfig) GetInts(path string, delim string) (value []int64, err error) {
	return value, c.GrabValues(path, delim,
		func(cap int) { value = make([]int64, 0, cap) },
		func(data interface{}) error {
			var parsed int64
			if parsed, err = parseYamlInt(data); err == nil {
				value = append(value, parsed)
			}
			return err
		})
}

// Get subconfig.
func (c *yamlConfig) GetConfigPart(path string) (Config, error) {
	element, err := c.findElement(path)
	if err != nil {
		return nil, err
	}
	return &yamlConfig{data: element}, nil
}

// Yaml helpers.
func (c *yamlConfig) findElement(path string) (interface{}, error) {
	element := c.data
	pathParts := splitPath(path)
	if len(pathParts) == 0 {
		return nil, ErrorNotFound
	}
	for _, pathPart := range pathParts {
		part, converted := element.(map[interface{}]interface{})
		if !converted {
			return nil, ErrorNotFound
		}
		var exist bool
		element, exist = part[pathPart]
		if !exist {
			return nil, ErrorNotFound
		}
	}
	return element, nil
}

// Yaml value parsers.
func parseYamlString(data interface{}) (value string, err error) {
	return parseJsonString(data)
}

func parseYamlBool(data interface{}) (value bool, err error) {
	return parseJsonBool(data)
}

func parseYamlFloat(data interface{}) (value float64, err error) {
	return parseJsonFloat(data)
}

func parseYamlInt(data interface{}) (value int64, err error) {
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