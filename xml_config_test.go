package config

import (
	"testing"
	"reflect"
	"fmt"
)

const (
	xmlConfigPart = `<stringElement>value</stringElement><boolElement>true</boolElement>
		<floatElement>1.23456</floatElement><intElement>123456</intElement>
		<stringElements>value1 value2 value3</stringElements>
		<boolElements>true false true</boolElements>
		<floatElements>1.23 4.56 7.89</floatElements>
		<intElements>123 456 789</intElements>`
)

var (
	oneLevelXmlConfig = fmt.Sprintf(`<xml>%s</xml>`, xmlConfigPart)
	twoLevelXmlConfig = fmt.Sprintf(`<xml><first>%[1]s</first><second>%[1]s</second></xml>`, xmlConfigPart)
	manyLevelXmlConfig = fmt.Sprintf(`<xml><root><child1>%[1]s</child1>
		<child><grandchild><first>%[1]s</first><second>%[1]s</second></grandchild></child></root>
		<root1><child>%[1]s</child></root1></xml>`, xmlConfigPart)
)

func equalXmlTest(t *testing.T, data string, path string, functors Functors) {
	config, err := newXmlConfig([]byte(data))
	if err != nil {
		t.Errorf("Cannot parse xml-config: %v", err)
		return
	}
	value, err := functors.Getter(config, path)
	if err != nil {
		t.Errorf("Cannot get value of '%s': %v", path, err)
		return
	}
	functors.Checker(t, value)
}

// Tests.
func TestCreateEmptyXml(t *testing.T) {
	_, err := newXmlConfig([]byte("<root/>"))
	if err != nil {
		t.Errorf("Cannot parse empty xml-config: %v", err)
	}
}

func TestOneLevelXml(t *testing.T) {
	for element, functors := range elementFunctors {
		equalXmlTest(t, oneLevelXmlConfig, joinPath("/xml", element), functors)
	}
}

func TestTwoLevelXml(t *testing.T) {
	for element, functors := range elementFunctors {
		equalXmlTest(t, twoLevelXmlConfig, joinPath("/xml/first", element), functors)
		equalXmlTest(t, twoLevelXmlConfig, joinPath("/xml/second", element), functors)
	}
}

func TestManyLevelXml(t *testing.T) {
	for element, functors := range elementFunctors {
		equalXmlTest(t, manyLevelXmlConfig, joinPath("/xml/root/child/grandchild/first", element), functors)
		equalXmlTest(t, manyLevelXmlConfig, joinPath("/xml/root/child/grandchild/second", element), functors)
	}
}

func TestManyLevelXmlGetValue(t *testing.T) {
	config, err := newXmlConfig([]byte(manyLevelXmlConfig))
	if err != nil {
		t.Errorf("Cannot parse xml-config: %v", err)
		return
	}

	value := configData{}
	err = LoadValue(config, "/xml/root/child/grandchild/first", &value)
	if err != nil {
		t.Errorf("Cannot load value from config: %v", err)
		return
	}

	checkStringValue(t, value.StringElement)
	checkBoolValue(t, value.BoolElement)
	checkFloatValue(t, value.FloatElement)
	checkIntValue(t, value.IntElement)

	checkStringValues(t, value.StringElements)
	checkBoolValues(t, value.BoolElements)
	checkFloatValues(t, value.FloatElements)
	checkIntValues(t, value.IntElements)
}

func TestManyLevelXmlGetConfigPart(t *testing.T) {
	rootConfig, err := newXmlConfig([]byte(manyLevelXmlConfig))
	if err != nil {
		t.Errorf("Cannot parse xml-config: %v", err)
		return
	}
	configPart, err := rootConfig.GetConfigPart("/xml/root/child/grandchild/first")
	if err != nil {
		t.Errorf("Cannot get config part: %v", err)
		return
	}

	expectedConfig, err := newXmlConfig([]byte(oneLevelXmlConfig))
	if err != nil {
		t.Errorf("Cannot parse expected json-config: %v", err)
		return
	}
	expectedConfigPart, err := expectedConfig.GetConfigPart("/xml")
	if err != nil {
		t.Errorf("Cannot get config part: %v", err)
		return
	}

	if !reflect.DeepEqual(configPart, expectedConfigPart) {
		t.Errorf("Not equal configs: expected - %v, actual - %v", expectedConfigPart, configPart)
		return
	}
}
