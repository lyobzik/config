package config

import (
	"testing"
	"fmt"

	"github.com/stretchr/testify/require"
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
	require.NoError(t, err, "Cannot parse xml-config")

	value, err := functors.Getter(config, path)
	require.NoError(t, err, "Cannot get value of '%s'", path)

	functors.Checker(t, value)
}

// Tests.
func TestCreateEmptyXml(t *testing.T) {
	_, err := newXmlConfig([]byte("<root/>"))
	require.NoError(t, err, "Cannot parse empty xml-config")
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

func TestManyLevelXmlLoadValue(t *testing.T) {
	config, err := newXmlConfig([]byte(manyLevelXmlConfig))
	require.NoError(t, err, "Cannot parse xml-config")

	value := configData{}
	err = LoadValueIgnoringErrors(config, "/xml/root/child/grandchild/first", &value)
	require.NoError(t, err, "Cannot load value from config")

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
	require.NoError(t, err, "Cannot parse xml-config")

	configPart, err := rootConfig.GetConfigPart("/xml/root/child/grandchild/first")
	require.NoError(t, err, "Cannot get config part")

	expectedConfig, err := newXmlConfig([]byte(oneLevelXmlConfig))
	require.NoError(t, err, "Cannot parse expected json-config")

	expectedConfigPart, err := expectedConfig.GetConfigPart("/xml")
	require.NoError(t, err, "Cannot get config part")

	require.Equal(t, configPart, expectedConfigPart, "Not equal configs")
}
