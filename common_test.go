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

// Test helpers for single value.
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

func getDurationValue(config Config, path string) (interface{}, error) {
	return GetDuration(config, path)
}

func checkDurationValue(t *testing.T, value interface{}) {
	expectedDuration, _ := time.ParseDuration("2h45m5s150ms")
	checkEqual(t, value, expectedDuration)
}

func getTimeValue(config Config, path string) (interface{}, error) {
	return GetTime(config, path)
}

func checkTimeValue(t *testing.T, value interface{}) {
	expectedTime, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05+07:00")
	checkEqual(t, value, expectedTime)
}

// Test helpers for array values.
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

func getDurationValues(config Config, path string) (interface{}, error) {
	return GetDurations(config, path, " ")
}

func checkDurationValues(t *testing.T, value interface{}) {
	stringDurations := []string{"1h", "1h15m30s450ms", "1s750ms"}
	durations := make([]time.Duration, len(stringDurations))
	for i := range stringDurations {
		durations[i], _ = time.ParseDuration(stringDurations[i])
	}
	checkEqual(t, value, durations)
}

func getTimeValues(config Config, path string) (interface{}, error) {
	return GetTimes(config, path, " ")
}

func checkTimeValues(t *testing.T, value interface{}) {
	stringTimes := []string{"2006-01-02T15:04:05+07:00", "2015-01-02T01:15:45Z", "1999-12-31T23:59:59+00:00"}
	times := make([]time.Time, len(stringTimes))
	for i := range stringTimes {
		times[i], _ = time.Parse(time.RFC3339, stringTimes[i])
	}
	checkEqual(t, value, times)
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
		"intElements": Functors{Getter: getIntValues, Checker: checkIntValues},
		"durationElement": Functors{Getter: getDurationValue, Checker: checkDurationValue},
		"timeElement": Functors{Getter: getTimeValue, Checker: checkTimeValue},
		"durationElements": Functors{Getter: getDurationValues, Checker: checkDurationValues},
		"timeElements": Functors{Getter: getTimeValues, Checker: checkTimeValues},
	}
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

	DurationElement time.Duration `config:"durationElement"`
	TimeElement time.Time `config:"timeElement"`
	DurationElements []time.Duration `config:"durationElements"`
	TimeElements []time.Time `config:"timeElements"`
}
