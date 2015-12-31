package config

 import (
	"encoding/json"
	"errors"
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

func (c *jsonConfig) GetValue(path string) (value interface{}, err error) {
	return c.FindElement(path)
}

func (c *jsonConfig) GetString(path string) (value string, err error) {
	element, err := c.FindElement(path)
	if err != nil {
		return value, err
	}
	if value, converted := element.(string); converted {
		return value, err
	}
	return value, errors.New("Value is not string")
}

func (c *jsonConfig) GetBool(path string) (value bool, err error) {
	element, err := c.FindElement(path)
	if err != nil {
		return value, err
	}
	if value, converted := element.(bool); converted {
		return value, err
	}
	return value, errors.New("Value is not bool")
}

func (c *jsonConfig) GetFloat(path string) (value float64, err error) {
	element, err := c.FindElement(path)
	if err != nil {
		return value, err
	}
	if value, converted := element.(float64); converted {
		return value, err
	}
	return value, errors.New("Value is not float64")
}

func (c *jsonConfig) GetInt(path string) (value int64, err error) {
	element, err := c.FindElement(path)
	if err != nil {
		return value, err
	}
	floatingValue, converted := element.(float64)
	if !converted {
		return value, errors.New("Value is not int")
	}
	// Check that value is integer.
	if math.Abs(math.Trunc(floatingValue) - floatingValue) < math.Nextafter(0, 1) {
		return int64(floatingValue), nil
	}
	return value, errors.New("Value is not int")
}

func (c *jsonConfig) GetConfigPart(path string) (Config, error) {
	element, err := c.FindElement(path)
	if err != nil {
		return nil, err
	}
	return &jsonConfig{data: element}, nil
}

func (c *jsonConfig) LoadValue(path string, value interface{}) (err error) {
	element, err := c.FindElement(path)
	if err != nil {
		return err
	}
	serializedElement, err := json.Marshal(element)
	if err != nil {
		return err
	}
	return json.Unmarshal(serializedElement, value)
}

func (c *jsonConfig) FindElement(path string) (interface{}, error) {
	var element interface{}
	element = c.data
	for _, pathPart := range splitPath(path) {
		if part, ok := element.(map[string]interface{}); ok {
			element = part[pathPart]
		} else {
			return nil, ErrorNotFound
		}
	}
	return element, nil
}