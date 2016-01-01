package config

import (
	"errors"
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

func (c *yamlConfig) GetValue(path string) (value interface{}, err error) {
	return c.FindElement(path)
}

func (c *yamlConfig) GetString(path string) (value string, err error) {
	element, err := c.FindElement(path)
	if err != nil {
		return value, err
	}
	if value, converted := element.(string); converted {
		return value, err
	}
	return value, errors.New("Value is not string")
}

func (c *yamlConfig) GetBool(path string) (value bool, err error) {
	element, err := c.FindElement(path)
	if err != nil {
		return value, err
	}
	if value, converted := element.(bool); converted {
		return value, err
	}
	return value, errors.New("Value is not bool")
}

func (c *yamlConfig) GetFloat(path string) (value float64, err error) {
	element, err := c.FindElement(path)
	if err != nil {
		return value, err
	}
	if value, converted := element.(float64); converted {
		return value, err
	}
	return value, errors.New("Value is not float64")
}

func (c *yamlConfig) GetInt(path string) (value int64, err error) {
	element, err := c.FindElement(path)
	if err != nil {
		return value, err
	}
	switch elementValue := element.(type) {
	case int:
		return int64(elementValue), nil
	case int64:
		return int64(elementValue), nil
	case float64:
		// Check that value is integer.
		if math.Abs(math.Trunc(elementValue) - elementValue) < math.Nextafter(0, 1) {
			return int64(elementValue), nil
		}
		return value, errors.New("Value is not int")
	}
	return value, errors.New("Value is not int")
}

func (c *yamlConfig) GetConfigPart(path string) (Config, error) {
	element, err := c.FindElement(path)
	if err != nil {
		return nil, err
	}
	return &yamlConfig{data: element}, nil
}

func (c *yamlConfig) LoadValue(path string, value interface{}) (err error) {
	element, err := c.FindElement(path)
	if err != nil {
		return err
	}
	serializedElement, err := yaml.Marshal(element)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(serializedElement, value)
}

func (c *yamlConfig) FindElement(path string) (interface{}, error) {
	var element interface{}
	element = c.data
	for _, pathPart := range splitPath(path) {
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