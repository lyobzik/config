package config

import (
	"reflect"
	"testing"
	"strings"
)

// Helpers to get and check values.
func checkEqual(t *testing.T, value, expected interface{}) {
	if !reflect.DeepEqual(value, expected) {
		t.Errorf("Incorrect value: expected - %v, actual - %v",  expected, value)
	}
}

func joinPath(pathParts ...string) string {
	path := strings.Join(pathParts, "/")
	if strings.HasPrefix(path, "/") {
		return path
	}
	return "/" + path
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

type Functors struct {
	Getter func(Config, string) (interface{}, error)
	Checker func(*testing.T, interface{})
}

var (
	elementFunctors = map[string]Functors{
		"stringElement": Functors{Getter: getStringValue, Checker: checkStringValue},
		"boolElement": Functors{Getter: getBoolValue, Checker: checkBoolValue},
		"floatElement": Functors{Getter: getFloatValue, Checker: checkFloatValue},
		"intElement": Functors{Getter: getIntValue, Checker: checkIntValue}}
)

// Settings structure used in tests.
type configData struct {
	StringElement string `ini:"stringElement"`
	BoolElement bool `ini:"boolElement"`
	FloatElement float64 `ini:"floatElement"`
	IntElement int64 `ini:"intElement"`
}
