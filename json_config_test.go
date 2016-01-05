package config

import (
	"testing"
	"reflect"
	"fmt"
)

var (
	oneLevelJsonConfig = `{"stringElement": "value", "boolElement": true,
		"floatElement": 1.23456, "intElement": 123456,
		"stringElements": ["value1", "value2", "value3"],
		"boolElements": [true, false, true],
		"floatElements": [1.23, 4.56, 7.89],
		"intElements": [123, 456, 789]}`
	twoLevelJsonConfig = fmt.Sprintf(`{"first": %[1]s, "second": %[1]s}`, oneLevelJsonConfig)
	manyLevelJsonConfig = fmt.Sprintf(`{"root": {"child1": %[1]s, "child": {"grandchild": %[1]s}},
		"root1": {"child": %[1]s}}`, twoLevelJsonConfig)
)

func equalJsonTest(t *testing.T, data string, path string, functors Functors) {
	config, err := newJsonConfig([]byte(data))
	if err != nil {
		t.Errorf("Cannot parse json-config: %v", err)
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
func TestCreateEmptyJson(t *testing.T) {
	_, err := newJsonConfig([]byte("{}"))
	if err != nil {
		t.Errorf("Cannot parse empty json-config: %v", err)
	}
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
	if err != nil {
		t.Errorf("Cannot parse json-config: %v", err)
		return
	}

	value := configData{}
	err = LoadValue(config, "/root/child/grandchild/first", &value)
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

func TestManyLevelJsonGetConfigPart(t *testing.T) {
	rootConfig, err := newJsonConfig([]byte(manyLevelJsonConfig))
	if err != nil {
		t.Errorf("Cannot parse root json-config: %v", err)
		return
	}
	expectedConfig, err := newJsonConfig([]byte(oneLevelJsonConfig))
	if err != nil {
		t.Errorf("Cannot parse expected json-config: %v", err)
		return
	}
	configPart, err := rootConfig.GetConfigPart("/root/child/grandchild/first")
	if err != nil {
		t.Errorf("Cannot get config part: %v", err)
		return
	}

	if !reflect.DeepEqual(configPart, expectedConfig) {
		t.Errorf("Not equal configs: expected - %v, actual - %v", expectedConfig, configPart)
		return
	}
}
