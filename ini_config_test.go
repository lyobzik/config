package config

import (
	"testing"
	"fmt"

	"github.com/stretchr/testify/require"
)

var (
	oneLevelIniConfig = "stringElement=value\nboolElement=true\n" +
		"floatElement=1.23456\nintElement=123456\n" +
		"stringElements=value1 value2 value3\nboolElements=true false true\n" +
		"floatElements=1.23 4.56 7.89\nintElements=123 456 789\n" +
		"timeElement=2006-01-02T15:04:05+07:00\ndurationElement=2h45m5s150ms\n" +
		"timeElements=2006-01-02T15:04:05+07:00 2015-01-02T01:15:45Z 1999-12-31T23:59:59+00:00\n" +
		"durationElements=1h 1h15m30s450ms 1s750ms"

	twoLevelIniConfig = fmt.Sprintf("[first]\n%[1]s\n[second]\n%[1]s", oneLevelIniConfig)
)

func equalIniTest(t *testing.T, data string, path string, functors Functors) {
	config, err := newIniConfig([]byte(data))
	require.Nil(t, err, "Cannot parse ini-config")

	value, err := functors.Getter(config, path)
	require.Nil(t, err, "Cannot get value of '%s'", path)

	functors.Checker(t, value)
}

// Tests.
func TestCreateEmptyIni(t *testing.T) {
	_, err := newIniConfig([]byte(""))
	require.Nil(t, err, "Cannot parse empty ini-config")
}

func TestOneLevelIni(t *testing.T) {
	for element, functors := range elementFunctors {
		equalIniTest(t, oneLevelIniConfig, element, functors)
	}
}

func TestTwoLevelIni(t *testing.T) {
	for element, functors := range elementFunctors {
		equalIniTest(t, twoLevelIniConfig, joinPath("first", element), functors)
		equalIniTest(t, twoLevelIniConfig, joinPath("second", element), functors)
	}
}

func TestTwoLevelIniLoadValue(t *testing.T) {
	config, err := newIniConfig([]byte(twoLevelIniConfig))
	require.Nil(t, err, "Cannot parse ini-config")

	value := configData{}
	err = LoadValueIgnoringErrors(config, "/first", &value)
	require.Nil(t, err, "Cannot load value from config")

	value.Check(t)
}

func TestTwoLevelIniGetConfigPart(t *testing.T) {
	expectedConfig, err := newIniConfig([]byte(oneLevelIniConfig))
	require.Nil(t, err, "Cannot parse expected ini-config")

	expectedValue := configData{}
	err = LoadValueIgnoringErrors(expectedConfig, "/", &expectedValue)
	require.Nil(t, err, "Cannot load value from expected ini-config")

	rootConfig, err := newIniConfig([]byte(twoLevelIniConfig))
	require.Nil(t, err, "Cannot parse root ini-config")

	configPart, err := rootConfig.GetConfigPart("/first")
	require.Nil(t, err, "Cannot get config part")

	value := configData{}
	err = LoadValueIgnoringErrors(configPart, "/", &value)
	require.Nil(t, err, "Cannot load value from ini-config")

	require.Equal(t, value, expectedValue, "Not equal configs")
}

func TestIniGetEmptyStrings(t *testing.T) {
	config, err := newIniConfig([]byte("element="))
	require.NoError(t, err, "Cannot parse ini-config")

	value, err := config.GetStrings("/element", DEFAULT_ARRAY_DELIMITER)
	require.NoError(t, err, "Cannot get value")

	require.Empty(t, value)
}

// Negative tests.
func TestIncorrectIniConfig(t *testing.T) {
	_, err := newIniConfig([]byte("{}"))
	require.Error(t, err, "Incorrect ini-config parsed successfully")
}

func TestIniGetValueEmptyPath(t *testing.T) {
	config, err := newIniConfig([]byte(""))
	require.NoError(t, err, "Cannot parse ini-config")

	_, err = config.GetString("")
	require.EqualError(t, err, ErrorNotFound.Error())

	_, err = config.GetStrings("", DEFAULT_ARRAY_DELIMITER)
	require.EqualError(t, err, ErrorNotFound.Error())
}

func TestEmptyIniGetAbsentValue(t *testing.T) {
	config, err := newIniConfig([]byte(""))
	require.NoError(t, err, "Cannot parse ini-config")

	_, err = config.GetString("/")
	require.Error(t, err, "Parameter must be absent")

	_, err = config.GetStrings("/", " ")
	require.Error(t, err, "Parameter must be absent")

	_, err = config.GetString("/element")
	require.Error(t, err, "Parameter must be absent")

	_, err = config.GetStrings("/element", " ")
	require.Error(t, err, "Parameter must be absent")

	_, err = config.GetString("/first/element")
	require.Error(t, err, "Parameter must be absent")

	_, err = config.GetStrings("/first/element", " ")
	require.Error(t, err, "Parameter must be absent")

	_, err = config.GetString("/first/child/element")
	require.Error(t, err, "Parameter must be absent")

	_, err = config.GetStrings("/first/child/element", " ")
	require.Error(t, err, "Parameter must be absent")
}

func TestOneLevelIniGetAbsentValue(t *testing.T) {
	config, err := newIniConfig([]byte(oneLevelIniConfig))
	require.NoError(t, err, "Cannot parse ini-config")

	_, err = config.GetString("/")
	require.Error(t, err, "Parameter must be absent")

	_, err = config.GetStrings("/", " ")
	require.Error(t, err, "Parameter must be absent")

	_, err = config.GetString("/element")
	require.Error(t, err, "Parameter must be absent")

	_, err = config.GetStrings("/element", " ")
	require.Error(t, err, "Parameter must be absent")

	_, err = config.GetString("/first/element")
	require.Error(t, err, "Parameter must be absent")

	_, err = config.GetStrings("/first/element", " ")
	require.Error(t, err, "Parameter must be absent")

	_, err = config.GetString("/first/child/element")
	require.Error(t, err, "Parameter must be absent")

	_, err = config.GetStrings("/first/child/element", " ")
	require.Error(t, err, "Parameter must be absent")
}

func TestTwoLevelIniGetAbsentValue(t *testing.T) {
	config, err := newIniConfig([]byte(twoLevelIniConfig))
	require.NoError(t, err, "Cannot parse ini-config")

	_, err = config.GetString("/")
	require.Error(t, err, "Parameter must be absent")

	_, err = config.GetStrings("/", " ")
	require.Error(t, err, "Parameter must be absent")

	_, err = config.GetString("/element")
	require.Error(t, err, "Parameter must be absent")

	_, err = config.GetStrings("/element", " ")
	require.Error(t, err, "Parameter must be absent")

	_, err = config.GetString("/first/element")
	require.Error(t, err, "Parameter must be absent")

	_, err = config.GetStrings("/first/element", " ")
	require.Error(t, err, "Parameter must be absent")

	_, err = config.GetString("/first/child/element")
	require.Error(t, err, "Parameter must be absent")

	_, err = config.GetStrings("/first/child/element", " ")
	require.Error(t, err, "Parameter must be absent")
}

func TestIniGetValueOfIncorrectType(t *testing.T) {
	config, err := newIniConfig([]byte(oneLevelIniConfig))
	require.NoError(t, err, "Cannot parse ini-config")

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
