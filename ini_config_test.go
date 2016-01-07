package config

import (
	"testing"
	"reflect"
	"fmt"
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
	if err != nil {
		t.Errorf("Cannot parse ini-config: %v", err)
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
func TestCreateEmptyIni(t *testing.T) {
	_, err := newIniConfig([]byte(""))
	if err != nil {
		t.Errorf("Cannot parse empty ini-config: %v", err)
	}
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
	if err != nil {
		t.Errorf("Cannot parse ini-config: %v", err)
		return
	}

	value := configData{}
	err = LoadValueIgnoringErrors(config, "/first", &value)
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

func TestTwoLevelIniGetConfigPart(t *testing.T) {
	expectedConfig, err := newIniConfig([]byte(oneLevelIniConfig))
	if err != nil {
		t.Errorf("Cannot parse expected ini-config: %v", err)
		return
	}
	expectedValue := configData{}
	err = LoadValueIgnoringErrors(expectedConfig, "/", &expectedValue)
	if err != nil {
		t.Errorf("Cannot load value from expected ini-config: %v", err)
		return
	}

	rootConfig, err := newIniConfig([]byte(twoLevelIniConfig))
	if err != nil {
		t.Errorf("Cannot parse root ini-config: %v", err)
		return
	}
	configPart, err := rootConfig.GetConfigPart("/first")
	if err != nil {
		t.Errorf("Cannot get config part: %v", err)
		return
	}
	value := configData{}
	err = LoadValueIgnoringErrors(configPart, "/", &value)
	if err != nil {
		t.Errorf("Cannot load value from ini-config: %v", err)
		return
	}

	if !reflect.DeepEqual(value, expectedValue) {
		t.Errorf("Not equal configs: expected - %v, actual - %v", expectedValue, value)
		return
	}
}
