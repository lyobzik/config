//Copyright 2016 lyobzik
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package config

import (
	"errors"
	"fmt"
	"testing"

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
	oneLevelXMLConfig  = fmt.Sprintf(`<xml>%s</xml>`, xmlConfigPart)
	twoLevelXMLConfig  = fmt.Sprintf(`<xml><first>%[1]s</first><second>%[1]s</second></xml>`, xmlConfigPart)
	manyLevelXMLConfig = fmt.Sprintf(`<xml><root><child1>%[1]s</child1>
		<child><grandchild><first>%[1]s</first><second>%[1]s</second></grandchild></child></root>
		<root1><child>%[1]s</child></root1></xml>`, xmlConfigPart)
)

func equalXMLTest(t *testing.T, data string, path string, functors Functors) {
	config, err := newXMLConfig([]byte(data))
	require.NoError(t, err, "Cannot parse xml-config")

	value, err := functors.Getter(config, path)
	require.NoError(t, err, "Cannot get value of '%s'", path)

	functors.Checker(t, value)
}

// Tests.
func TestCreateEmptyXml(t *testing.T) {
	_, err := newXMLConfig([]byte("<root/>"))
	require.NoError(t, err, "Cannot parse empty xml-config")
}

func TestOneLevelXml(t *testing.T) {
	for element, functors := range elementFunctors {
		equalXMLTest(t, oneLevelXMLConfig, joinPath("/xml", element), functors)
	}
}

func TestTwoLevelXml(t *testing.T) {
	for element, functors := range elementFunctors {
		equalXMLTest(t, twoLevelXMLConfig, joinPath("/xml/first", element), functors)
		equalXMLTest(t, twoLevelXMLConfig, joinPath("/xml/second", element), functors)
	}
}

func TestManyLevelXml(t *testing.T) {
	for element, functors := range elementFunctors {
		equalXMLTest(t, manyLevelXMLConfig, joinPath("/xml/root/child/grandchild/first", element), functors)
		equalXMLTest(t, manyLevelXMLConfig, joinPath("/xml/root/child/grandchild/second", element), functors)
	}
}

func TestManyLevelXmlLoadValue(t *testing.T) {
	config, err := newXMLConfig([]byte(manyLevelXMLConfig))
	require.NoError(t, err, "Cannot parse xml-config")

	value := configData{}
	err = LoadValueIgnoringMissingFieldErrors(config, "/xml/root/child/grandchild/first", &value)
	require.NoError(t, err, "Cannot load value from config")

	value.Check(t)
}

func TestXmlGetAttributeValue(t *testing.T) {
	config, err := newXMLConfig([]byte(`<xml element="value"/>`))
	require.NoError(t, err, "Cannot parse xml-config")

	value, err := config.GetString("/xml/@element")
	require.NoError(t, err, "Cannot get attribute value")

	require.Equal(t, value, "value")
}

func TestXmlGetEmptyStrings(t *testing.T) {
	config, err := newXMLConfig([]byte(`<xml/>`))
	require.NoError(t, err, "Cannot parse xml-config")

	value, err := config.GetStrings("/xml", defaultArrayDelimiter)
	require.NoError(t, err, "Cannot get value")

	require.Empty(t, value)
}

func TestXmlGrabValue(t *testing.T) {
	config, err := newXMLConfig([]byte(oneLevelXMLConfig))
	require.NoError(t, err, "Cannot parse xml-config")

	var intValue int64
	var convertingError error
	var isStringData bool
	err = config.GrabValue("/xml/intElement", func(data interface{}) error {
		var stringData string
		if stringData, isStringData = data.(string); !isStringData {
			return errors.New("Incorrect data type")
		}
		intValue, convertingError = parseXMLInt(stringData)
		return nil
	})

	require.NoError(t, err, "Cannot grab value from xml-config")
	require.True(t, isStringData, "Data must be string")
	require.NoError(t, convertingError, "Cannot convert intElement to int")
	checkIntValue(t, intValue)
}

func TestXmlGrabValues(t *testing.T) {
	config, err := newXMLConfig([]byte(oneLevelXMLConfig))
	require.NoError(t, err, "Cannot parse xml-config")

	var intValues []int64
	var isStringData bool
	err = config.GrabValues("/xml/intElements", defaultArrayDelimiter,
		func(length int) { intValues = make([]int64, 0, length) },
		func(data interface{}) error {
			var stringData string
			if stringData, isStringData = data.(string); !isStringData {
				return errors.New("Incorrect data type")
			}
			value, err := parseXMLInt(stringData)
			if err != nil {
				return err
			}
			intValues = append(intValues, value)
			return nil
		})

	require.NoError(t, err, "Cannot grab value from xml-config")
	require.True(t, isStringData, "Data must be string")
	checkIntValues(t, intValues)
}

func TestXmlGrabValuesOfSingleElement(t *testing.T) {
	config, err := newXMLConfig([]byte(oneLevelXMLConfig))
	require.NoError(t, err, "Cannot parse xml-config")

	var intValues []int64
	var isStringData bool
	err = config.GrabValues("/xml/intElement", defaultArrayDelimiter,
		func(length int) { intValues = make([]int64, 0, length) },
		func(data interface{}) error {
			var stringData string
			if stringData, isStringData = data.(string); !isStringData {
				return errors.New("Incorrect data type")
			}
			value, err := parseXMLInt(stringData)
			if err != nil {
				return err
			}
			intValues = append(intValues, value)
			return nil
		})

	require.NoError(t, err, "Cannot grab value from xml-config")
	require.True(t, isStringData, "Data must be string")
	require.Len(t, intValues, 1)
	checkIntValue(t, intValues[0])
}

// Negative tests.
func TestIncorrectXmlConfig(t *testing.T) {
	_, err := newXMLConfig([]byte("<root"))
	require.Error(t, err, "Incorrect xml-config parsed successfully")
}

func TestXmlGetValueEmptyPath(t *testing.T) {
	config, err := newXMLConfig([]byte(`<xml/>`))
	require.NoError(t, err, "Cannot parse xml-config")

	for _, functors := range elementFunctors {
		_, err = functors.Getter(config, "")
		require.EqualError(t, err, ErrorNotFound.Error())
	}
}

func TestXmlGetAbsentValue(t *testing.T) {
	config, err := newXMLConfig([]byte(`<xml/>`))
	require.NoError(t, err, "Cannot parse xml-config")

	for _, functors := range elementFunctors {
		_, err = functors.Getter(config, "/xml/element")
		require.Error(t, err, "Attribute must be absent")
	}
}

func TestXmlGetAbsentAttributeValue(t *testing.T) {
	config, err := newXMLConfig([]byte(`<xml/>`))
	require.NoError(t, err, "Cannot parse xml-config")

	for _, functors := range elementFunctors {
		_, err = functors.Getter(config, "/xml/@element")
		require.Error(t, err, "Attribute must be absent")
	}
}

func TestXmlGetValueOfIncorrectType(t *testing.T) {
	config, err := newXMLConfig([]byte(oneLevelXMLConfig))
	require.NoError(t, err, "Cannot parse xml-config")

	_, err = config.GetBool("/xml/stringElement")
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetInt("/xml/stringElement")
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetFloat("/xml/stringElement")
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetBools("/xml/stringElement", defaultArrayDelimiter)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetInts("/xml/stringElement", defaultArrayDelimiter)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetFloats("/xml/stringElement", defaultArrayDelimiter)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetBools("/xml/stringElements", defaultArrayDelimiter)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetInts("/xml/stringElements", defaultArrayDelimiter)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetFloats("/xml/stringElements", defaultArrayDelimiter)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetInt("/xml/floatElement")
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetInts("/xml/floatElements", defaultArrayDelimiter)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")
}

func TestXmlGrabAbsentValue(t *testing.T) {
	config, err := newXMLConfig([]byte(oneLevelXMLConfig))
	require.NoError(t, err, "Cannot parse xml-config")

	executed := false
	err = config.GrabValue("/xml/absentElement", func(data interface{}) error {
		executed = true
		return nil
	})

	require.EqualError(t, err, ErrorNotFound.Error())
	require.Equal(t, false, executed, "Value grabber must not be executed")
}

func TestXmlGrabAbsentValues(t *testing.T) {
	config, err := newXMLConfig([]byte(oneLevelXMLConfig))
	require.NoError(t, err, "Cannot parse xml-config")

	executed := false
	err = config.GrabValues("/xml/absentElement", defaultArrayDelimiter,
		func(length int) { executed = true },
		func(data interface{}) error {
			executed = true
			return nil
		})

	require.EqualError(t, err, ErrorNotFound.Error())
	require.Equal(t, false, executed, "Value grabber must not be executed")
}

func TestXmlGrabValuePassError(t *testing.T) {
	config, err := newXMLConfig([]byte(oneLevelXMLConfig))
	require.NoError(t, err, "Cannot parse xml-config")

	expectedError := errors.New("TestXmlGrabValuePassError error")
	err = config.GrabValue("/xml/intElement", func(data interface{}) error {
		return expectedError
	})

	require.EqualError(t, err, expectedError.Error())
}

func TestXmlGrabValuesPassError(t *testing.T) {
	config, err := newXMLConfig([]byte(oneLevelXMLConfig))
	require.NoError(t, err, "Cannot parse xml-config")

	expectedError := errors.New("TestXmlGrabValuesPassError error")
	err = config.GrabValues("/xml/intElements", defaultArrayDelimiter,
		func(length int) {},
		func(data interface{}) error {
			return expectedError
		})

	require.EqualError(t, err, expectedError.Error())
}

// Parser tests.
func TestParseXmlBool(t *testing.T) {
	value, err := parseXMLBool(fmt.Sprint(expectedBoolValue))
	require.NoError(t, err, "Cannot parse xml bool")
	checkBoolValue(t, value)

	_, err = parseXMLBool(expectedStringValue)
	require.EqualError(t, err, ErrorIncorrectValueType.Error())
}

func TestParseXmlFloat(t *testing.T) {
	value, err := parseXMLFloat(fmt.Sprint(expectedFloatValue))
	require.NoError(t, err, "Cannot parse xml float")
	checkFloatValue(t, value)

	_, err = parseXMLFloat(expectedStringValue)
	require.EqualError(t, err, ErrorIncorrectValueType.Error())
}

func TestParseXmlInt(t *testing.T) {
	value, err := parseXMLInt(fmt.Sprint(expectedIntValue))
	require.NoError(t, err, "Cannot parse xml int")
	checkIntValue(t, value)

	value, err = parseXMLInt(fmt.Sprint(float64(expectedIntValue)))
	require.NoError(t, err, "Cannot parse xml int")
	checkIntValue(t, value)

	_, err = parseXMLInt(fmt.Sprint(float64(expectedIntValue) + 0.00000001))
	require.EqualError(t, err, ErrorIncorrectValueType.Error())

	_, err = parseXMLInt(expectedStringValue)
	require.EqualError(t, err, ErrorIncorrectValueType.Error())
}

// Test GetConfigPart
func TestXmlGetConfigPartRootFromRoot(t *testing.T) {
	rootConfig, err := newXMLConfig([]byte(twoLevelXMLConfig))
	require.Nil(t, err, "Cannot parse root xml-config")

	configPart, err := rootConfig.GetConfigPart("/")
	require.Nil(t, err, "Cannot get config part")

	require.Equal(t, rootConfig, configPart, "Not equal configs")
}

func TestXmlGetConfigPartSectionFromRoot(t *testing.T) {
	expectedConfig, err := newXMLConfig([]byte(oneLevelXMLConfig))
	require.NoError(t, err, "Cannot parse expected xml-config")

	expectedConfigPart, err := expectedConfig.GetConfigPart("/xml")
	require.NoError(t, err, "Cannot get config part")

	rootConfig, err := newXMLConfig([]byte(manyLevelXMLConfig))
	require.NoError(t, err, "Cannot parse root xml-config")

	configPart, err := rootConfig.GetConfigPart("/xml/root/child/grandchild/first")
	require.NoError(t, err, "Cannot get config part")

	require.Equal(t, expectedConfigPart, configPart, "Not equal configs")
}

func TestXmlGetConfigPartSectionFromSection(t *testing.T) {
	rootConfig, err := newXMLConfig([]byte(manyLevelXMLConfig))
	require.NoError(t, err, "Cannot parse root xml-config")

	configSection, err := rootConfig.GetConfigPart("/xml/root/child/grandchild/first")
	require.NoError(t, err, "Cannot get config section")

	configPart, err := configSection.GetConfigPart("/")
	require.NoError(t, err, "Cannot get config part")

	require.Equal(t, configSection, configPart, "Not equal configs")
}

func TestXmlGetConfigPartWithLongPath(t *testing.T) {
	rootConfig, err := newXMLConfig([]byte(manyLevelXMLConfig))
	require.NoError(t, err, "Cannot parse root xml-config")

	configSection, err := rootConfig.GetConfigPart("/xml/root/child/grandchild")
	require.NoError(t, err, "Cannot get config section")

	_, err = rootConfig.GetConfigPart("/xml/root/child/grandchild/first/stringElement/element")
	require.EqualError(t, err, ErrorNotFound.Error())

	_, err = configSection.GetConfigPart("/first/stringElement/element")
	require.EqualError(t, err, ErrorNotFound.Error())
}

func TestXmlGetAbsentConfigPart(t *testing.T) {
	config, err := newXMLConfig([]byte(`<xml element="asd"/>`))
	require.NoError(t, err, "Cannot parse xml-config")

	_, err = config.GetConfigPart("/root")
	require.Error(t, err, ErrorNotFound.Error())

	_, err = config.GetConfigPart("/xml/@element")
	require.Error(t, err, ErrorNotFound.Error())
}
