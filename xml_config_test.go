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
		<intElements>123 456 789</intElements>
		<timeElement>2006-01-02T15:04:05+07:00</timeElement>
		<durationElement>2h45m5s150ms</durationElement>
		<timeElements>2006-01-02T15:04:05+07:00 2015-01-02T01:15:45Z 1999-12-31T23:59:59+00:00</timeElements>
		<durationElements>1h 1h15m30s450ms 1s750ms</durationElements>`
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

	value.Check(t)
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

func TestXmlGetAttributeValue(t *testing.T) {
	config, err := newXmlConfig([]byte(`<xml element="value"/>`))
	require.NoError(t, err, "Cannot parse xml-config")

	value, err := config.GetString("/xml/@element")
	require.NoError(t, err, "Cannot get attribute value")

	require.Equal(t, value, "value")
}

func TestXmlGetEmptyStrings(t *testing.T) {
	config, err := newXmlConfig([]byte(`<xml/>`))
	require.NoError(t, err, "Cannot parse xml-config")

	value, err := config.GetStrings("/xml", " ")
	require.NoError(t, err, "Cannot get value")

	require.Empty(t, value)
}

// Negative tests.
func TestIncorrectXmlConfig(t *testing.T) {
	_, err := newXmlConfig([]byte("<root"))
	require.Error(t, err, "Incorrect xml-config parsed successfully")
}

func TestXmlGetAbsentValue(t *testing.T) {
	config, err := newXmlConfig([]byte(`<xml/>`))
	require.NoError(t, err, "Cannot parse xml-config")

	_, err = config.GetString("/xml/element")
	require.Error(t, err, "Attribute must be absent")

	_, err = config.GetStrings("/xml/element", " ")
	require.Error(t, err, "Attribute must be absent")

}

func TestXmlGetAbsentAttributeValue(t *testing.T) {
	config, err := newXmlConfig([]byte(`<xml/>`))
	require.NoError(t, err, "Cannot parse xml-config")

	_, err = config.GetString("/xml/@element")
	require.Error(t, err, "Attribute must be absent")
}

func TestXmlGetValueOfIncorrectType(t *testing.T) {
	config, err := newXmlConfig([]byte(oneLevelXmlConfig))
	require.NoError(t, err, "Cannot parse xml-config")

	_, err = config.GetBool("/xml/stringElement")
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetInt("/xml/stringElement")
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetFloat("/xml/stringElement")
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetBools("/xml/stringElement", DEFAULT_ARRAY_DELIMITER)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetInts("/xml/stringElement", DEFAULT_ARRAY_DELIMITER)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetFloats("/xml/stringElement", DEFAULT_ARRAY_DELIMITER)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetBools("/xml/stringElements", DEFAULT_ARRAY_DELIMITER)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetInts("/xml/stringElements", DEFAULT_ARRAY_DELIMITER)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetFloats("/xml/stringElements", DEFAULT_ARRAY_DELIMITER)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")
}
