package config

import (
	"testing"
	"strings"
	"fmt"

	"github.com/stretchr/testify/require"
)

var (
	oneLevelYamlConfig = "stringElement: value\nboolElement: true\n" +
		"floatElement: 1.23456\nintElement: 123456\n" +
		"stringElements: [value1, value2, value3]\nboolElements: [true, false, true]\n" +
		"floatElements: [1.23, 4.56, 7.89]\nintElements: [123, 456, 789]"
	twoLevelYamlConfig = fmt.Sprintf("first: %[1]s\nsecond: %[1]s",
		strings.Replace("\n" + oneLevelYamlConfig, "\n", "\n  ", -1))
	manyLevelYamlConfig = fmt.Sprintf("root:\n  child1: %[1]s\n  child:\n    grandchild: %[2]s\n" +
		"root1:\n  child: %[1]s",
		strings.Replace("\n" + twoLevelYamlConfig, "\n", "\n    ", -1),
		strings.Replace("\n" + twoLevelYamlConfig, "\n", "\n      ", -1))
)

func equalYamlTest(t *testing.T, data string, path string, functors Functors) {
	config, err := newYamlConfig([]byte(data))
	require.NoError(t, err, "Cannot parse yaml-config")

	value, err := functors.Getter(config, path)
	require.NoError(t, err, "Cannot get value of '%s'", path)

	functors.Checker(t, value)
}

// Tests.
func TestCreateEmptyYaml(t *testing.T) {
	_, err := newJsonConfig([]byte("{}"))
	require.NoError(t, err, "Cannot parse empty yaml-config")
}

func TestOneLevelYaml(t *testing.T) {
	for element, functors := range elementFunctors {
		equalYamlTest(t, oneLevelYamlConfig, element, functors)
	}
}

func TestTwoLevelYaml(t *testing.T) {
	for element, functors := range elementFunctors {
		equalYamlTest(t, twoLevelYamlConfig, joinPath("first", element), functors)
		equalYamlTest(t, twoLevelYamlConfig, joinPath("second", element), functors)
	}
}

func TestManyLevelYaml(t *testing.T) {
	for element, functors := range elementFunctors {
		equalYamlTest(t, manyLevelYamlConfig, joinPath("/root/child/grandchild/first", element), functors)
		equalYamlTest(t, manyLevelYamlConfig, joinPath("/root/child/grandchild/second", element), functors)
	}
}

func TestManyLevelYamlLoadValue(t *testing.T) {
	config, err := newYamlConfig([]byte(manyLevelYamlConfig))
	require.NoError(t, err, "Cannot parse yaml-config")

	value := configData{}
	err = LoadValueIgnoringErrors(config, "/root/child/grandchild/first", &value)
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

func TestManyLevelYamlGetConfigPart(t *testing.T) {
	rootConfig, err := newYamlConfig([]byte(manyLevelYamlConfig))
	require.NoError(t, err, "Cannot parse root yaml-config")

	expectedConfig, err := newYamlConfig([]byte(oneLevelYamlConfig))
	require.NoError(t, err, "Cannot parse expected yaml-config")

	configPart, err := rootConfig.GetConfigPart("/root/child/grandchild/first")
	require.NoError(t, err, "Cannot get config part")

	require.Equal(t, configPart, expectedConfig, "Not equal configs")
}
