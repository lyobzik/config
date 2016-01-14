package config

import (
	"errors"
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
		"timeElement": "2006-01-02T15:04:05+07:00",
		"durationElement": "2h45m5s150ms",
		"timeElements": ["2006-01-02T15:04:05+07:00", "2015-01-02T01:15:45Z", "1999-12-31T23:59:59+00:00"],
		"durationElements": ["1h", "1h15m30s450ms", "1s750ms"]}`
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
	err = LoadValueIgnoringErrors(config, "/root/child/grandchild/first", &value)
	require.NoError(t, err, "Cannot load value from config")

	value.Check(t)
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

func TestJsonGetEmptyStrings(t *testing.T) {
	config, err := newJsonConfig([]byte(`{"stringElements": []}`))
	require.NoError(t, err, "Cannot parse json-config")

	value, err := config.GetStrings("/stringElements", " ")
	require.NoError(t, err, "Cannot get value")

	require.Empty(t, value)
}

func TestJsonGetFloatAsInt(t *testing.T) {
	config, err := newJsonConfig([]byte(`{"intElement": 1.0, "intElements": [1.0, 2.0, 3.0]}`))
	require.NoError(t, err, "Cannot parse json-config")

	intValue, err := config.GetInt("/intElement")
	require.NoError(t, err, "Cannot get value")
	require.Equal(t, intValue, int64(1))

	intValues, err := config.GetInts("/intElements", " ")
	require.NoError(t, err, "Cannot get value")
	require.Equal(t, intValues, []int64{1, 2, 3})
}

func TestJsonGrabValue(t *testing.T) {
	config, err := newJsonConfig([]byte(oneLevelJsonConfig))
	require.NoError(t, err, "Cannot parse json-config")

	var intValue int64
	var convertingError error
	err = config.GrabValue("/intElement", func(data interface{}) error {
		intValue, convertingError = parseJsonInt(data)
		return nil
	})

	require.NoError(t, err, "Cannot grab value from json-config")
	require.NoError(t, convertingError, "Cannot convert intElement to int")
	checkIntValue(t, intValue)
}

func TestJsonGrabValues(t *testing.T) {
	config, err := newJsonConfig([]byte(oneLevelJsonConfig))
	require.NoError(t, err, "Cannot parse json-config")

	var intValues []int64
	err = config.GrabValues("/intElements", DEFAULT_ARRAY_DELIMITER,
		func(length int) { intValues = make([]int64, 0, length) },
		func(data interface{}) error {
			value, err := parseJsonInt(data)
			if err != nil {
				return err
			}
			intValues = append(intValues, value)
			return nil
		})

	require.NoError(t, err, "Cannot grab value from json-config")
	checkIntValues(t, intValues)
}

// Negative tests.
func TestIncorrectJsonConfig(t *testing.T) {
	_, err := newJsonConfig([]byte("{"))
	require.Error(t, err, "Incorrect json-config parsed successfully")
}

func TestJsonGetValueEmptyPath(t *testing.T) {
	config, err := newJsonConfig([]byte(`{"element": "value"}`))
	require.NoError(t, err, "Cannot parse json-config")

	_, err = config.GetString("")
	require.EqualError(t, err, ErrorNotFound.Error())

	_, err = config.GetStrings("", DEFAULT_ARRAY_DELIMITER)
	require.EqualError(t, err, ErrorNotFound.Error())
}

func TestJsonGetAbsentValue(t *testing.T) {
	config, err := newJsonConfig([]byte(`{"element": "value"}`))
	require.NoError(t, err, "Cannot parse json-config")

	_, err = config.GetString("/root")
	require.Error(t, err, ErrorNotFound.Error())

	_, err = config.GetStrings("/root", DEFAULT_ARRAY_DELIMITER)
	require.Error(t, err, ErrorNotFound.Error())
}

func TestJsonGetAbsentConfigPart(t *testing.T) {
	config, err := newJsonConfig([]byte(`{"element": "value"}`))
	require.NoError(t, err, "Cannot parse json-config")

	_, err = config.GetConfigPart("/root")
	require.Error(t, err, ErrorNotFound.Error())
}

func TestJsonGetValueOfIncorrectType(t *testing.T) {
	config, err := newJsonConfig([]byte(oneLevelJsonConfig))
	require.NoError(t, err, "Cannot parse json-config")

	_, err = config.GetString("/intElement")
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetBool("/stringElement")
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetInt("/stringElement")
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetFloat("/stringElement")
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetStrings("/intElement", DEFAULT_ARRAY_DELIMITER)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetBools("/stringElement", DEFAULT_ARRAY_DELIMITER)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetInts("/stringElement", DEFAULT_ARRAY_DELIMITER)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetFloats("/stringElement", DEFAULT_ARRAY_DELIMITER)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetStrings("/intElements", DEFAULT_ARRAY_DELIMITER)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetBools("/stringElements", DEFAULT_ARRAY_DELIMITER)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetInts("/stringElements", DEFAULT_ARRAY_DELIMITER)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetFloats("/stringElements", DEFAULT_ARRAY_DELIMITER)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetInt("/floatElement")
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetInts("/floatElements", DEFAULT_ARRAY_DELIMITER)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")
}

func TestJsonGrabAbsentValue(t *testing.T) {
	config, err := newJsonConfig([]byte(oneLevelJsonConfig))
	require.NoError(t, err, "Cannot parse json-config")

	executed := false
	err = config.GrabValue("/absentElement", func(data interface{}) error {
		executed = true
		return nil
	})

	require.EqualError(t, err, ErrorNotFound.Error())
	require.Equal(t, false, executed, "Value grabber must not be executed")
}

func TestJsonGrabAbsentValues(t *testing.T) {
	config, err := newJsonConfig([]byte(oneLevelJsonConfig))
	require.NoError(t, err, "Cannot parse json-config")

	executed := false
	err = config.GrabValues("/absentElement", DEFAULT_ARRAY_DELIMITER,
		func(length int) {executed = true},
		func(data interface{}) error {
			executed = true
			return nil
		})

	require.EqualError(t, err, ErrorNotFound.Error())
	require.Equal(t, false, executed, "Value grabber must not be executed")
}

func TestJsonGrabValuePassError(t *testing.T) {
	config, err := newJsonConfig([]byte(oneLevelJsonConfig))
	require.NoError(t, err, "Cannot parse json-config")

	expectedError := errors.New("TestJsonGrabValuePassError error")
	err = config.GrabValue("/intElement", func(data interface{}) error {
		return expectedError
	})

	require.EqualError(t, err, expectedError.Error())
}

func TestJsonGrabValuesPassError(t *testing.T) {
	config, err := newJsonConfig([]byte(oneLevelJsonConfig))
	require.NoError(t, err, "Cannot parse json-config")

	expectedError := errors.New("TestJsonGrabValuesPassError error")
	err = config.GrabValues("/intElements", DEFAULT_ARRAY_DELIMITER,
		func(length int) {},
		func(data interface{}) error {
			return expectedError
		})

	require.EqualError(t, err, expectedError.Error())
}

func TestJsonGrabValuesOfSingleElement(t *testing.T) {
	config, err := newJsonConfig([]byte(oneLevelJsonConfig))
	require.NoError(t, err, "Cannot parse json-config")

	err = config.GrabValues("/intElement", DEFAULT_ARRAY_DELIMITER,
		func(length int) {},
		func(data interface{}) error {
			return nil
		})

	require.EqualError(t, err, ErrorIncorrectValueType.Error())
}
