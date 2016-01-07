package config

import (
	"reflect"
	"testing"
	"time"
)

// Helpers to get and check values.
func checkEqual(t *testing.T, value, expected interface{}) {
	if !reflect.DeepEqual(value, expected) {
		t.Errorf("Incorrect value: expected - %v, actual - %v",  expected, value)
	}
}

func getStringValue(config Config, path string) (interface{}, error) {
	return config.GetString(path)
}

func checkStringValue(t *testing.T, value interface{}) {
	checkEqual(t, value, "value")
}

func getBoolValue(config Config, path string) (interface{}, error) {
	return config.GetBool(path)
}

func checkBoolValue(t *testing.T, value interface{}) {
	checkEqual(t, value, true)
}

func getFloatValue(config Config, path string) (interface{}, error) {
	return config.GetFloat(path)
}

func checkFloatValue(t *testing.T, value interface{}) {
	checkEqual(t, value, 1.23456)
}

func getIntValue(config Config, path string) (interface{}, error) {
	return config.GetInt(path)
}

func checkIntValue(t *testing.T, value interface{}) {
	checkEqual(t, value, int64(123456))
}

func getStringValues(config Config, path string) (interface{}, error) {
	return config.GetStrings(path, " ")
}

func checkStringValues(t *testing.T, value interface{}) {
	checkEqual(t, value, []string{"value1", "value2", "value3"})
}

func getBoolValues(config Config, path string) (interface{}, error) {
	return config.GetBools(path, " ")
}

func checkBoolValues(t *testing.T, value interface{}) {
	checkEqual(t, value, []bool{true, false, true})
}

func getFloatValues(config Config, path string) (interface{}, error) {
	return config.GetFloats(path, " ")
}

func checkFloatValues(t *testing.T, value interface{}) {
	checkEqual(t, value, []float64{1.23, 4.56, 7.89})
}

func getIntValues(config Config, path string) (interface{}, error) {
	return config.GetInts(path, " ")
}

func checkIntValues(t *testing.T, value interface{}) {
	checkEqual(t, value, []int64{123, 456, 789})
}

type Functors struct {
	Getter func(Config, string) (interface{}, error)
	Checker func(*testing.T, interface{})
}

var (
	elementFunctors = map[string]Functors{
		"stringElement": Functors{Getter: getStringValue, Checker: checkStringValue},
		"boolElement": Functors{Getter: getBoolValue, Checker: checkBoolValue},
		"floatElement": Functors{Getter: getFloatValue, Checker: checkFloatValue},
		"intElement": Functors{Getter: getIntValue, Checker: checkIntValue},
		"stringElements": Functors{Getter: getStringValues, Checker: checkStringValues},
		"boolElements": Functors{Getter: getBoolValues, Checker: checkBoolValues},
		"floatElements": Functors{Getter: getFloatValues, Checker: checkFloatValues},
		"intElements": Functors{Getter: getIntValues, Checker: checkIntValues}}
)

// Settings structure used in tests.
type configData struct {
	StringElement string `config:"stringElement"`
	BoolElement bool `config:"boolElement"`
	FloatElement float64 `config:"floatElement"`
	IntElement int64 `config:"intElement"`

	StringElements []string `config:"stringElements"`
	BoolElements []bool `config:"boolElements"`
	FloatElements []float64 `config:"floatElements"`
	IntElements []int64 `config:"intElements"`

	TimeElement time.Time
}
