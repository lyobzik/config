package config

import (
	"testing"
	"reflect"
	"fmt"
)

var (
	oneLevelJsonConfig = `{"stringElement": "value", "boolElement": true,
		"floatElement": 1.23456, "intElement": 123456}`
	twoLevelJsonConfig = fmt.Sprintf(`{"first": %[1]s, "second": %[1]s}`, oneLevelJsonConfig)
	manyLevelJsonConfig = fmt.Sprintf(`{"root": {"child1": %[1]s, "child": {"grandchild": %[1]s}},
		"root1": {"child": %[1]s}}`, twoLevelJsonConfig)
)

func equalTest(t *testing.T, data string, path string, functors Functors) {
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
func TestCreateEmpty(t *testing.T) {
	_, err := newJsonConfig([]byte("{}"))
	if err != nil {
		t.Errorf("Cannot parse empty json-config: %v", err)
	}
}

func TestOneLevel(t *testing.T) {
	for element, functors := range elementFunctors {
		equalTest(t, oneLevelJsonConfig, element, functors)
	}
}

func TestTwoLevel(t *testing.T) {
	for element, functors := range elementFunctors {
		equalTest(t, twoLevelJsonConfig, joinPath("first", element), functors)
		equalTest(t, twoLevelJsonConfig, joinPath("second", element), functors)
	}
}

func TestManyLevel(t *testing.T) {
	for element, functors := range elementFunctors {
		equalTest(t, manyLevelJsonConfig, joinPath("/root/child/grandchild/first", element), functors)
		equalTest(t, manyLevelJsonConfig, joinPath("/root/child/grandchild/second", element), functors)
	}
}

func TestManyLevelGetValue(t *testing.T) {
	config, err := newJsonConfig([]byte(manyLevelJsonConfig))
	if err != nil {
		t.Errorf("Cannot parse root json-config: %v", err)
		return
	}

	value := configData{}
	err = config.LoadValue("/root/child/grandchild/first", &value)
	if err != nil {
		t.Errorf("Cannot load value from config: %v", err)
		return
	}

	checkStringValue(t, value.StringElement)
	checkBoolValue(t, value.BoolElement)
	checkFloatValue(t, value.FloatElement)
	checkIntValue(t, value.IntElement)
}

func TestManyLevelConfigPart(t *testing.T) {
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
