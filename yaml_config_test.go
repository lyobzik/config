package config

import (
	"reflect"
	"testing"
	"strings"
	"fmt"
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
	if err != nil {
		t.Errorf("Cannot parse yaml-config: %v", err)
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
func TestCreateEmptyYaml(t *testing.T) {
	_, err := newJsonConfig([]byte("{}"))
	if err != nil {
		t.Errorf("Cannot parse empty yaml-config: %v", err)
	}
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

func TestManyLevelYamlGetValue(t *testing.T) {
	config, err := newYamlConfig([]byte(manyLevelYamlConfig))
	if err != nil {
		t.Errorf("Cannot parse yaml-config: %v", err)
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

func TestManyLevelYamlGetConfigPart(t *testing.T) {
	rootConfig, err := newYamlConfig([]byte(manyLevelYamlConfig))
	if err != nil {
		t.Errorf("Cannot parse root yaml-config: %v", err)
		return
	}
	expectedConfig, err := newYamlConfig([]byte(oneLevelYamlConfig))
	if err != nil {
		t.Errorf("Cannot parse expected yaml-config: %v", err)
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
