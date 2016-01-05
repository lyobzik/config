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

func (c *yamlConfig) GetType() string {
	return YAML
}

func (c *yamlConfig) GetValue(path string) (value interface{}, err error) {
	return c.FindElement(path)
}

func (c *yamlConfig) GetString(path string) (value string, err error) {
	element, err := c.FindElement(path)
	if err != nil {
		return value, err
	}
	return parseYamlString(element)
}

func (c *yamlConfig) GetBool(path string) (value bool, err error) {
	element, err := c.FindElement(path)
	if err != nil {
		return value, err
	}
	return parseYamlBool(element)
}

func (c *yamlConfig) GetFloat(path string) (value float64, err error) {
	element, err := c.FindElement(path)
	if err != nil {
		return value, err
	}
	return parseYamlFloat(element)
}

func (c *yamlConfig) GetInt(path string) (value int64, err error) {
	element, err := c.FindElement(path)
	if err != nil {
		return value, err
	}
	return parseYamlInt(element)
}

func (c *yamlConfig) GetStrings(path string, delim string) (value []string, err error) {
	element, err := c.FindElement(path)
	if err != nil {
		return value, err
	}
	arrayValue, converted := element.([]interface{})
	if !converted {
		return value, ErrorIncorrectValueType
	}
	resultValue := make([]string, len(arrayValue))
	for i := range arrayValue {
		if resultValue[i], err = parseYamlString(arrayValue[i]); err != nil {
			return value, err
		}
	}
	return resultValue, nil
}

func (c *yamlConfig) GetBools(path string, delim string) (value []bool, err error) {
	element, err := c.FindElement(path)
	if err != nil {
		return value, err
	}
	arrayValue, converted := element.([]interface{})
	if !converted {
		return value, ErrorIncorrectValueType
	}
	resultValue := make([]bool, len(arrayValue))
	for i := range arrayValue {
		if resultValue[i], err = parseYamlBool(arrayValue[i]); err != nil {
			return value, err
		}
	}
	return resultValue, nil
}

func (c *yamlConfig) GetFloats(path string, delim string) (value []float64, err error) {
	element, err := c.FindElement(path)
	if err != nil {
		return value, err
	}
	arrayValue, converted := element.([]interface{})
	if !converted {
		return value, ErrorIncorrectValueType
	}
	resultValue := make([]float64, len(arrayValue))
	for i := range arrayValue {
		if resultValue[i], err = parseYamlFloat(arrayValue[i]); err != nil {
			return value, err
		}
	}
	return resultValue, nil
}

func (c *yamlConfig) GetInts(path string, delim string) (value []int64, err error) {
	element, err := c.FindElement(path)
	if err != nil {
		return value, err
	}
	arrayValue, converted := element.([]interface{})
	if !converted {
		return value, ErrorIncorrectValueType
	}
	resultValue := make([]int64, len(arrayValue))
	for i := range arrayValue {
		if resultValue[i], err = parseYamlInt(arrayValue[i]); err != nil {
			return value, err
		}
	}
	return resultValue, nil
}

func (c *yamlConfig) GetConfigPart(path string) (Config, error) {
	element, err := c.FindElement(path)
	if err != nil {
		return nil, err
	}
	return &yamlConfig{data: element}, nil
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

// Helpers.
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