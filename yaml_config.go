package config

import (
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
	element, err := c.findElement(path)
	if err != nil {
		return err
	}
	return grabber(element)
}

func (c *yamlConfig) GrabValues(path string, delim string,
	creator ValueSliceCreator, grabber ValueGrabber) (err error) {

	return c.GrabValue(path, createYamlValueGrabber(creator, grabber))
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
		if element, exist = part[pathPart]; !exist {
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
	return parseJsonInt(data)
}

// Grabbing helpers.
func createYamlValueGrabber(creator ValueSliceCreator, grabber ValueGrabber) ValueGrabber {
	return createJsonValueGrabber(creator, grabber)
}
