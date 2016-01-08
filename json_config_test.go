package config

import (
	"testing"
	"fmt"

	"github.com/stretchr/testify/require"
)

var (
	oneLevelJsonConfig = `{"stringElement": "value", "boolElement": true,
		"floatElement": 1.23456, "intElement": 123456,
		"stringElements": ["value1", "value2", "value3"],
		"boolElements": [true, false, true],
		"floatElements": [1.23, 4.56, 7.89],
		"intElements": [123, 456, 789],
		"TimeElement": "2006-01-02T15:04:05+07:00"}`
	twoLevelJsonConfig = fmt.Sprintf(`{"first": %[1]s, "second": %[1]s}`, oneLevelJsonConfig)
	manyLevelJsonConfig = fmt.Sprintf(`{"root": {"child1": %[1]s, "child": {"grandchild": %[1]s}},
		"root1": {"child": %[1]s}}`, twoLevelJsonConfig)
)

func equalJsonTest(t *testing.T, data string, path string, functors Functors) {
	config, err := newJsonConfig([]byte(data))
	require.NoError(t, err, "Cannot parse json-config")

	value, err := functors.Getter(config, path)
	require.NoError(t, err, "Cannot get value of '%s'", path)

	functors.Checker(t, value)
}

// Tests.
func TestCreateEmptyJson(t *testing.T) {
	_, err := newJsonConfig([]byte("{}"))
	require.NoError(t, err, "Cannot parse empty json-config")
}

func TestOneLevelJson(t *testing.T) {
	for element, functors := range elementFunctors {
		equalJsonTest(t, oneLevelJsonConfig, element, functors)
	}
}

func TestTwoLevelJson(t *testing.T) {
	for element, functors := range elementFunctors {
		equalJsonTest(t, twoLevelJsonConfig, joinPath("first", element), functors)
		equalJsonTest(t, twoLevelJsonConfig, joinPath("second", element), functors)
	}
}

func TestManyLevelJson(t *testing.T) {
	for element, functors := range elementFunctors {
		equalJsonTest(t, manyLevelJsonConfig, joinPath("/root/child/grandchild/first", element), functors)
		equalJsonTest(t, manyLevelJsonConfig, joinPath("/root/child/grandchild/second", element), functors)
	}
}

func TestManyLevelJsonLoadValue(t *testing.T) {
	config, err := newJsonConfig([]byte(manyLevelJsonConfig))
	require.NoError(t, err, "Cannot parse json-config")

	value := configData{}
	err = LoadValue(config, "/root/child/grandchild/first", &value)
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

func TestManyLevelJsonGetConfigPart(t *testing.T) {
	rootConfig, err := newJsonConfig([]byte(manyLevelJsonConfig))
	require.NoError(t, err, "Cannot parse root json-config")

	expectedConfig, err := newJsonConfig([]byte(oneLevelJsonConfig))
	require.NoError(t, err, "Cannot parse expected json-config")

	configPart, err := rootConfig.GetConfigPart("/root/child/grandchild/first")
	require.NoError(t, err, "Cannot get config part")

	require.Equal(t, configPart, expectedConfig, "Not equal configs")
}
