package config

import (
	"bytes"
	"encoding/xml"
	"io"
	"strconv"
	"strings"
	"unicode"
)

// Simple xml parsing.
type xmlElement struct {
	Attributes map[string]string
	Value      string
	Children   map[string][]*xmlElement
}

func newXMLElement() *xmlElement {
	return &xmlElement{
		Attributes: make(map[string]string),
		Value:      "",
		Children:   make(map[string][]*xmlElement, 0)}
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

func parseXML(data []byte) (*xmlElement, error) {
	reader := bytes.NewReader(data)
	decoder := xml.NewDecoder(reader)

	xmlRoot := newXMLElement()
	elements := []*xmlElement{xmlRoot}
	lastElement := func() *xmlElement {
		return elements[len(elements)-1]
	}

	token, err := decoder.Token()
	for ; err == nil; token, err = decoder.Token() {
		switch element := token.(type) {
		case xml.StartElement:
			newElement := newXMLElement()
			newElement.SetAttributes(element.Attr)
			lastElement().AddChild(element.Name.Local, newElement)

			elements = append(elements, newElement)
		case xml.EndElement:
			elements = elements[:len(elements)-1]
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

func newXMLConfig(data []byte) (Config, error) {
	xmlRoot, err := parseXML(data)
	if err != nil {
		return nil, err
	}
	return &xmlConfig{data: xmlRoot}, nil
}

// Grabbers.
func (c *xmlConfig) GrabValue(path string, grabber ValueGrabber) (err error) {
	return GrabStringValue(c, path, createXMLValueGrabber(grabber))
}

func (c *xmlConfig) GrabValues(path string, delim string,
	creator ValueSliceCreator, grabber ValueGrabber) (err error) {

	return GrabStringValues(c, path, delim, creator, createXMLValueGrabber(grabber))
}

// Get single value.
func (c *xmlConfig) GetString(path string) (value string, err error) {
	element, attribute, err := c.findElement(path)
	if err != nil {
		return "", err
	}
	if element != nil {
		return element.Value, nil
	}
	return attribute, nil
}

func (c *xmlConfig) GetBool(path string) (value bool, err error) {
	return value, GrabStringValue(c, path, func(data string) error {
		value, err = parseXMLBool(data)
		return err
	})
}

func (c *xmlConfig) GetFloat(path string) (value float64, err error) {
	return value, GrabStringValue(c, path, func(data string) error {
		value, err = parseXMLFloat(data)
		return err
	})
}

func (c *xmlConfig) GetInt(path string) (value int64, err error) {
	return value, GrabStringValue(c, path, func(data string) error {
		value, err = parseXMLInt(data)
		return err
	})
}

// Get array of values.
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
	return value, GrabStringValues(c, path, delim,
		func(cap int) { value = make([]bool, 0, cap) },
		func(data string) error {
			var parsed bool
			if parsed, err = parseXMLBool(data); err == nil {
				value = append(value, parsed)
			}
			return err
		})
}

func (c *xmlConfig) GetFloats(path string, delim string) (value []float64, err error) {
	return value, GrabStringValues(c, path, delim,
		func(cap int) { value = make([]float64, 0, cap) },
		func(data string) error {
			var parsed float64
			if parsed, err = parseXMLFloat(data); err == nil {
				value = append(value, parsed)
			}
			return err
		})
}

func (c *xmlConfig) GetInts(path string, delim string) (value []int64, err error) {
	return value, GrabStringValues(c, path, delim,
		func(cap int) { value = make([]int64, 0, cap) },
		func(data string) error {
			var parsed int64
			if parsed, err = parseXMLInt(data); err == nil {
				value = append(value, parsed)
			}
			return err
		})
}

// Get subconfig.
func (c *xmlConfig) GetConfigPart(path string) (Config, error) {
	if len(splitPath(path)) == 0 {
		return c, nil
	}
	element, _, err := c.findElement(path)
	if err != nil {
		return nil, err
	}
	if element == nil {
		return nil, ErrorNotFound
	}
	return &xmlConfig{data: element}, nil
}

// Xml helpers.
func (c *xmlConfig) findElement(path string) (*xmlElement, string, error) {
	element := c.data
	pathParts := splitPath(path)
	if len(pathParts) == 0 {
		return nil, "", ErrorNotFound
	}
	for _, pathPart := range pathParts {
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

// Xml value parsers.
func parseXMLBool(data string) (value bool, err error) {
	value, err = strconv.ParseBool(data)
	if err != nil {
		return false, ErrorIncorrectValueType
	}
	return value, nil
}

func parseXMLFloat(data string) (value float64, err error) {
	value, err = strconv.ParseFloat(data, 64)
	if err != nil {
		return 0.0, ErrorIncorrectValueType
	}
	return value, nil
}

func parseXMLInt(data string) (value int64, err error) {
	value, err = strconv.ParseInt(data, 10, 64)
	if err != nil {
		return 0, ErrorIncorrectValueType
	}
	return value, nil
}

// Grabbing helpers.
func createXMLValueGrabber(grabber ValueGrabber) StringValueGrabber {
	return func(data string) error {
		return grabber(data)
	}
}
