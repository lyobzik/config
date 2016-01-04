package config

import (
	"encoding/xml"
	"strings"
	"bytes"
	"unicode"
	"io"
	"strconv"
)

// Simple xml parsing.
type xmlElement struct {
	Attributes map[string]string
	Value string
	Children map[string][]*xmlElement
}

func NewXmlElement() *xmlElement {
	return &xmlElement{
		Attributes: make(map[string]string),
		Value: "",
		Children: make(map[string][]*xmlElement, 0)}
}

func (e *xmlElement) SetAttributes(attibutes []xml.Attr) {
	for _, attribute := range attibutes {
		e.Attributes[attribute.Name.Local] = attribute.Value
	}
}

func (e *xmlElement) AddChild(name string, child *xmlElement) {
	e.Children[name] = append(e.Children[name], child)
}

func (e *xmlElement) SetValue(value string) {
	value = strings.TrimFunc(value, unicode.IsSpace)
	if len(value) > 0 {
		e.Value = value
	}
}

func parseXml(data[] byte) (*xmlElement, error) {
	reader := bytes.NewReader(data)
	decoder := xml.NewDecoder(reader)

	xmlRoot := NewXmlElement()
	elements := []*xmlElement{xmlRoot}
	lastElement := func () *xmlElement {
		return elements[len(elements) - 1]
	}

	token, err := decoder.Token()
	for ; err == nil; token, err = decoder.Token() {
		switch element := token.(type) {
		case xml.StartElement:
			newElement := NewXmlElement()
			newElement.SetAttributes(element.Attr)
			lastElement().AddChild(element.Name.Local, newElement)

			elements = append(elements, newElement)
		case xml.EndElement:
			elements = elements[:len(elements) - 1]
		case xml.CharData:
			lastElement().SetValue(string(element))
		}
	}

	if err != nil && err != io.EOF {
		return nil, err
	}

	return xmlRoot, nil
}

// Xml config implementation.
type xmlConfig struct {
	data *xmlElement
}

func newXmlConfig(data []byte) (Config, error) {
	xmlRoot, err := parseXml(data)
	if err != nil {
		return nil, err
	}
	return &xmlConfig{data: xmlRoot}, nil
}

func (c *xmlConfig) GetValue(path string) (value interface{}, err error) {
	return c.GetString(path)
}

func (c *xmlConfig) GetString(path string) (value string, err error) {
	element, attribute, err := c.FindElement(path)
	if err != nil {
		return "", nil
	}
	if element != nil {
		return element.Value, nil
	}
	return attribute, nil
}

func (c *xmlConfig) GetBool(path string) (value bool, err error) {
	stringValue, err := c.GetString(path)
	if err != nil {
		return false, err
	}
	return parseXmlBool(stringValue)
}

func (c *xmlConfig) GetFloat(path string) (value float64, err error) {
	stringValue, err := c.GetString(path)
	if err != nil {
		return 0.0, err
	}
	return parseXmlFloat(stringValue)
}

func (c *xmlConfig) GetInt(path string) (value int64, err error) {
	stringValue, err := c.GetString(path)
	if err != nil {
		return 0, err
	}
	return parseXmlInt(stringValue)
}

func (c *xmlConfig) GetStrings(path string, delim string) (value []string, err error) {
	stringValue, err := c.GetString(path)
	if err != nil {
		return value, err
	}
	if len(stringValue) == 0 {
		return make([]string, 0), nil
	}
	return strings.Split(stringValue, delim), nil
}

func (c *xmlConfig) GetBools(path string, delim string) (value []bool, err error) {
	stringValues, err := c.GetStrings(path, delim)
	if err != nil {
		return value, err
	}
	resultValue := make([]bool, len(stringValues))
	for i := range stringValues {
		if resultValue[i], err = parseXmlBool(stringValues[i]); err != nil {
			return value, err
		}
	}
	return resultValue, nil
}

func (c *xmlConfig) GetFloats(path string, delim string) (value []float64, err error) {
	stringValues, err := c.GetStrings(path, delim)
	if err != nil {
		return value, err
	}
	resultValue := make([]float64, len(stringValues))
	for i := range stringValues {
		if resultValue[i], err = parseXmlFloat(stringValues[i]); err != nil {
			return value, err
		}
	}
	return resultValue, nil
}

func (c *xmlConfig) GetInts(path string, delim string) (value []int64, err error) {
	stringValues, err := c.GetStrings(path, delim)
	if err != nil {
		return value, err
	}
	resultValue := make([]int64, len(stringValues))
	for i := range stringValues {
		if resultValue[i], err = parseXmlInt(stringValues[i]); err != nil {
			return value, err
		}
	}
	return resultValue, nil
}

func (c *xmlConfig) GetConfigPart(path string) (Config, error) {
	element, _, err := c.FindElement(path)
	if err != nil {
		return nil, err
	}
	if element == nil {
		return nil, ErrorNotFound
	}
	return &xmlConfig{data: element}, nil
}

func (c *xmlConfig) LoadValue(path string, value interface{}) (err error) {
	return ErrorNotFound
//	element, err := c.FindElement(path)
//	if err != nil {
//		return err
//	}
//	serializedElement, err := xml.Marshal(element)
//	if err != nil {
//		return err
//	}
//	return xml.Unmarshal(serializedElement, value)
}

func (c *xmlConfig) FindElement(path string) (*xmlElement, string, error) {
	var element *xmlElement
	element = c.data
	for _, pathPart := range splitPath(path) {
		if strings.HasPrefix(pathPart, "@") {
			if attribute, exist := element.Attributes[pathPart[1:]]; exist {
				return nil, attribute, nil
			}
			return nil, "", ErrorNotFound
		}
		if part, ok := element.Children[pathPart]; ok {
			element = part[0]
		} else {
			return nil, "", ErrorNotFound
		}
	}
	return element, "", nil
}

// Helpers.
func parseXmlBool(data string) (value bool, err error) {
	value, err = strconv.ParseBool(data)
	if err != nil {
		return false, ErrorIncorrectValueType
	}
	return value, nil
}

func parseXmlFloat(data string) (value float64, err error) {
	value, err = strconv.ParseFloat(data, 64)
	if err != nil {
		return 0.0, ErrorIncorrectValueType
	}
	return value, nil
}

func parseXmlInt(data string) (value int64, err error) {
	value, err = strconv.ParseInt(data, 10, 64)
	if err != nil {
		return 0, ErrorIncorrectValueType
	}
	return value, nil
}