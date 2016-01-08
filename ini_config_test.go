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
		"floatElements=1.23 4.56 7.89\nintElements=123 456 789"
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

	checkStringValue(t, value.StringElement)
	checkBoolValue(t, value.BoolElement)
	checkFloatValue(t, value.FloatElement)
	checkIntValue(t, value.IntElement)

	checkStringValues(t, value.StringElements)
	checkBoolValues(t, value.BoolElements)
	checkFloatValues(t, value.FloatElements)
	checkIntValues(t, value.IntElements)
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
