package config

import (
	yaml "gopkg.in/yaml.v2"
)

type yamlConfig struct {
	data interface{}
}

func newYAMLConfig(data []byte) (Config, error) {
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

	return c.GrabValue(path, createYAMLValueGrabber(creator, grabber))
}

// Get single value.
func (c *yamlConfig) GetString(path string) (value string, err error) {
	return value, c.GrabValue(path, func(data interface{}) error {
		value, err = parseYAMLString(data)
		return err
	})
}

func (c *yamlConfig) GetBool(path string) (value bool, err error) {
	return value, c.GrabValue(path, func(data interface{}) error {
		value, err = parseYAMLBool(data)
		return err
	})
}

func (c *yamlConfig) GetFloat(path string) (value float64, err error) {
	return value, c.GrabValue(path, func(data interface{}) error {
		value, err = parseYAMLFloat(data)
		return err
	})
}

func (c *yamlConfig) GetInt(path string) (value int64, err error) {
	return value, c.GrabValue(path, func(data interface{}) error {
		value, err = parseYAMLInt(data)
		return err
	})
}

// Get array of values.
func (c *yamlConfig) GetStrings(path string, delim string) (value []string, err error) {
	return value, c.GrabValues(path, delim,
		func(cap int) { value = make([]string, 0, cap) },
		func(data interface{}) error {
			var parsed string
			if parsed, err = parseYAMLString(data); err == nil {
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
			if parsed, err = parseYAMLBool(data); err == nil {
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
			if parsed, err = parseYAMLFloat(data); err == nil {
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
			if parsed, err = parseYAMLInt(data); err == nil {
				value = append(value, parsed)
			}
			return err
		})
}

// Get subconfig.
func (c *yamlConfig) GetConfigPart(path string) (Config, error) {
	if len(splitPath(path)) == 0 {
		return c, nil
	}
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
func parseYAMLString(data interface{}) (value string, err error) {
	return parseJSONString(data)
}

func parseYAMLBool(data interface{}) (value bool, err error) {
	return parseJSONBool(data)
}

func parseYAMLFloat(data interface{}) (value float64, err error) {
	return parseJSONFloat(data)
}

func parseYAMLInt(data interface{}) (value int64, err error) {
	return parseJSONInt(data)
}

// Grabbing helpers.
func createYAMLValueGrabber(creator ValueSliceCreator, grabber ValueGrabber) ValueGrabber {
	return createJSONValueGrabber(creator, grabber)
}
