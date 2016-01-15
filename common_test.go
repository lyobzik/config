package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	expectedStringValue = "value"
	expectedBoolValue = true
	expectedFloatValue = 1.23456
	expectedIntValue = int64(123456)

	expectedDurationValue = parseExpectedDurations("2h45m5s150ms")[0]
	expectedTimeValue = parseExpectedTimes("2006-01-02T15:04:05+07:00")[0]

	expectedStringValues = []string{"value1", "value2", "value3"}
	expectedBoolValues = []bool{true, false, true}
	expectedFloatValues = []float64{1.23, 4.56, 7.89}
	expectedIntValues = []int64{123, 456, 789}

	expectedDurationValues = parseExpectedDurations("1h", "1h15m30s450ms", "1s750ms")
	expectedTimeValues = parseExpectedTimes("2006-01-02T15:04:05+07:00", "2015-01-02T01:15:45Z",
		"1999-12-31T23:59:59+00:00")
)

// Helpers to get and check values.
func parseExpectedDurations(values... string) []time.Duration {
	result := make([]time.Duration, 0)
	for _, value := range values {
		duration, _ := time.ParseDuration(value)
		result = append(result, duration)
	}
	return result
}

func parseExpectedTimes(values... string) []time.Time {
	result := make([]time.Time, 0)
	for _, value := range values {
		time, _ := time.Parse(time.RFC3339, value)
		result = append(result, time)
	}
	return result
}

func checkEqual(t *testing.T, value, expected interface{}) {
	require.Equal(t, expected, value)
}

// Test helpers for single value.
func getStringValue(config Config, path string) (interface{}, error) {
	return config.GetString(path)
}

func checkStringValue(t *testing.T, value interface{}) {
	checkEqual(t, value, expectedStringValue)
}

func getBoolValue(config Config, path string) (interface{}, error) {
	return config.GetBool(path)
}

func checkBoolValue(t *testing.T, value interface{}) {
	checkEqual(t, value, expectedBoolValue)
}

func getFloatValue(config Config, path string) (interface{}, error) {
	return config.GetFloat(path)
}

func checkFloatValue(t *testing.T, value interface{}) {
	checkEqual(t, value, expectedFloatValue)
}

func getIntValue(config Config, path string) (interface{}, error) {
	return config.GetInt(path)
}

func checkIntValue(t *testing.T, value interface{}) {
	checkEqual(t, value, expectedIntValue)
}

func getDurationValue(config Config, path string) (interface{}, error) {
	return GetDuration(config, path)
}

func checkDurationValue(t *testing.T, value interface{}) {
	checkEqual(t, value, expectedDurationValue)
}

func getTimeValue(config Config, path string) (interface{}, error) {
	return GetTime(config, path)
}

func checkTimeValue(t *testing.T, value interface{}) {
	checkEqual(t, value, expectedTimeValue)
}

// Test helpers for array values.
func getStringValues(config Config, path string) (interface{}, error) {
	return config.GetStrings(path, DEFAULT_ARRAY_DELIMITER)
}

func checkStringValues(t *testing.T, value interface{}) {
	checkEqual(t, value, expectedStringValues)
}

func getBoolValues(config Config, path string) (interface{}, error) {
	return config.GetBools(path, DEFAULT_ARRAY_DELIMITER)
}

func checkBoolValues(t *testing.T, value interface{}) {
	checkEqual(t, value, expectedBoolValues)
}

func getFloatValues(config Config, path string) (interface{}, error) {
	return config.GetFloats(path, DEFAULT_ARRAY_DELIMITER)
}

func checkFloatValues(t *testing.T, value interface{}) {
	checkEqual(t, value, expectedFloatValues)
}

func getIntValues(config Config, path string) (interface{}, error) {
	return config.GetInts(path, DEFAULT_ARRAY_DELIMITER)
}

func checkIntValues(t *testing.T, value interface{}) {
	checkEqual(t, value, expectedIntValues)
}

func getDurationValues(config Config, path string) (interface{}, error) {
	return GetDurations(config, path, DEFAULT_ARRAY_DELIMITER)
}

func checkDurationValues(t *testing.T, value interface{}) {
	checkEqual(t, value, expectedDurationValues)
}

func getTimeValues(config Config, path string) (interface{}, error) {
	return GetTimes(config, path, DEFAULT_ARRAY_DELIMITER)
}

func checkTimeValues(t *testing.T, value interface{}) {
	checkEqual(t, value, expectedTimeValues)
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

func (data configData) Check(t *testing.T) {
	checkStringValue(t, data.StringElement)
	checkBoolValue(t, data.BoolElement)
	checkFloatValue(t, data.FloatElement)
	checkIntValue(t, data.IntElement)
	checkDurationValue(t, data.DurationElement)
	checkTimeValue(t, data.TimeElement)

	checkStringValues(t, data.StringElements)
	checkBoolValues(t, data.BoolElements)
	checkFloatValues(t, data.FloatElements)
	checkIntValues(t, data.IntElements)
	checkDurationValues(t, data.DurationElements)
	checkTimeValues(t, data.TimeElements)
}
