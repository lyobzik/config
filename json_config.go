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

func (c *jsonConfig) GetType() string {
	return JSON
}

func (c *jsonConfig) GetValue(path string) (value interface{}, err error) {
	return c.FindElement(path)
}

func (c *jsonConfig) GetString(path string) (value string, err error) {
	element, err := c.FindElement(path)
	if err != nil {
		return value, err
	}
	return parseJsonString(element)
}

func (c *jsonConfig) GetBool(path string) (value bool, err error) {
	element, err := c.FindElement(path)
	if err != nil {
		return value, err
	}
	return parseJsonBool(element)
}

func (c *jsonConfig) GetFloat(path string) (value float64, err error) {
	element, err := c.FindElement(path)
	if err != nil {
		return value, err
	}
	return parseJsonFloat(element)
}

func (c *jsonConfig) GetInt(path string) (value int64, err error) {
	element, err := c.FindElement(path)
	if err != nil {
		return value, err
	}
	return parseJsonInt(element)
}

func (c *jsonConfig) GetStrings(path string, delim string) (value []string, err error) {
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
		if resultValue[i], err = parseJsonString(arrayValue[i]); err != nil {
			return value, err
		}
	}
	return resultValue, nil
}

func (c *jsonConfig) GetBools(path string, delim string) (value []bool, err error) {
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
		if resultValue[i], err = parseJsonBool(arrayValue[i]); err != nil {
			return value, err
		}
	}
	return resultValue, nil
}

func (c *jsonConfig) GetFloats(path string, delim string) (value []float64, err error) {
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
		if resultValue[i], err = parseJsonFloat(arrayValue[i]); err != nil {
			return value, err
		}
	}
	return resultValue, nil
}

func (c *jsonConfig) GetInts(path string, delim string) (value []int64, err error) {
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
		if resultValue[i], err = parseJsonInt(arrayValue[i]); err != nil {
			return value, err
		}
	}
	return resultValue, nil
}

func (c *jsonConfig) GetConfigPart(path string) (Config, error) {
	element, err := c.FindElement(path)
	if err != nil {
		return nil, err
	}
	return &jsonConfig{data: element}, nil
}

func (c *jsonConfig) FindElement(path string) (interface{}, error) {
	var element interface{}
	element = c.data
	for _, pathPart := range splitPath(path) {
		part, converted := element.(map[string]interface{})
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
	floatingValue, converted := data.(float64)
	if !converted {
		return value, ErrorIncorrectValueType
	}
	// Check that value is integer.
	if math.Abs(math.Trunc(floatingValue) - floatingValue) < math.Nextafter(0, 1) {
		return int64(floatingValue), nil
	}
	return value, ErrorIncorrectValueType
}
