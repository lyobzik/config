package config

import (
	"errors"
	"testing"
	"strings"
	"fmt"

	"github.com/stretchr/testify/require"
)

var (
	oneLevelYamlConfig = "stringElement: value\nboolElement: true\n" +
		"floatElement: 1.23456\nintElement: 123456\n" +
		"stringElements: [value1, value2, value3]\nboolElements: [true, false, true]\n" +
		"floatElements: [1.23, 4.56, 7.89]\nintElements: [123, 456, 789]\n" +
		"timeElement: 2006-01-02T15:04:05+07:00\ndurationElement: 2h45m5s150ms\n" +
		"timeElements: [\"2006-01-02T15:04:05+07:00\", \"2015-01-02T01:15:45Z\", \"1999-12-31T23:59:59+00:00\"]\n" +
		"durationElements: [1h, 1h15m30s450ms, 1s750ms]"
	twoLevelYamlConfig = fmt.Sprintf("first: %[1]s\nsecond: %[1]s",
		strings.Replace("\n" + oneLevelYamlConfig, "\n", "\n  ", -1))
	manyLevelYamlConfig = fmt.Sprintf("root:\n  child1: %[1]s\n  child:\n    grandchild: %[2]s\n" +
		"root1:\n  child: %[1]s",
		strings.Replace("\n" + twoLevelYamlConfig, "\n", "\n    ", -1),
		strings.Replace("\n" + twoLevelYamlConfig, "\n", "\n      ", -1))
)

func equalYamlTest(t *testing.T, data string, path string, functors Functors) {
	config, err := newYamlConfig([]byte(data))
	require.NoError(t, err, "Cannot parse yaml-config")

	value, err := functors.Getter(config, path)
	require.NoError(t, err, "Cannot get value of '%s'", path)

	functors.Checker(t, value)
}

// Tests.
func TestCreateEmptyYaml(t *testing.T) {
	_, err := newJsonConfig([]byte("{}"))
	require.NoError(t, err, "Cannot parse empty yaml-config")
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

func TestManyLevelYamlLoadValue(t *testing.T) {
	config, err := newYamlConfig([]byte(manyLevelYamlConfig))
	require.NoError(t, err, "Cannot parse yaml-config")

	value := configData{}
	err = LoadValueIgnoringErrors(config, "/root/child/grandchild/first", &value)
	require.NoError(t, err, "Cannot load value from config")

	value.Check(t)
}

func TestManyLevelYamlGetConfigPart(t *testing.T) {
	rootConfig, err := newYamlConfig([]byte(manyLevelYamlConfig))
	require.NoError(t, err, "Cannot parse root yaml-config")

	expectedConfig, err := newYamlConfig([]byte(oneLevelYamlConfig))
	require.NoError(t, err, "Cannot parse expected yaml-config")

	configPart, err := rootConfig.GetConfigPart("/root/child/grandchild/first")
	require.NoError(t, err, "Cannot get config part")

	require.Equal(t, configPart, expectedConfig, "Not equal configs")
}

func TestYamlGetEmptyStrings(t *testing.T) {
	config, err := newYamlConfig([]byte("stringElements: []"))
	require.NoError(t, err, "Cannot parse yaml-config")

	value, err := config.GetStrings("/stringElements", " ")
	require.NoError(t, err, "Cannot get value")

	require.Empty(t, value)
}

func TestYamlGetFloatAsInt(t *testing.T) {
	config, err := newYamlConfig([]byte("intElement: 1.0\nintElements: [1.0, 2.0, 3.0]"))
	require.NoError(t, err, "Cannot parse yaml-config")

	intValue, err := config.GetInt("/intElement")
	require.NoError(t, err, "Cannot get value")
	require.Equal(t, intValue, int64(1))

	intValues, err := config.GetInts("/intElements", " ")
	require.NoError(t, err, "Cannot get value")
	require.Equal(t, intValues, []int64{1, 2, 3})
}

func TestYamlGrabValue(t *testing.T) {
	config, err := newYamlConfig([]byte(oneLevelYamlConfig))
	require.NoError(t, err, "Cannot parse yaml-config")

	var intValue int64
	var convertingError error
	err = config.GrabValue("/intElement", func(data interface{}) error {
		intValue, convertingError = parseYamlInt(data)
		return nil
	})

	require.NoError(t, err, "Cannot grab value from yaml-config")
	require.NoError(t, convertingError, "Cannot convert intElement to int")
	checkIntValue(t, intValue)
}

func TestYamlGrabValues(t *testing.T) {
	config, err := newYamlConfig([]byte(oneLevelYamlConfig))
	require.NoError(t, err, "Cannot parse yaml-config")

	var intValues []int64
	err = config.GrabValues("/intElements", DEFAULT_ARRAY_DELIMITER,
		func(length int) { intValues = make([]int64, 0, length) },
		func(data interface{}) error {
			value, err := parseYamlInt(data)
			if err != nil {
				return err
			}
			intValues = append(intValues, value)
			return nil
		})

	require.NoError(t, err, "Cannot grab value from yaml-config")
	checkIntValues(t, intValues)
}

// Negative tests.
func TestIncorrectYamlConfig(t *testing.T) {
	_, err := newYamlConfig([]byte("{"))
	require.Error(t, err, "Incorrect yaml-config parsed successfully")
}

func TestYamlGetValueEmptyPath(t *testing.T) {
	config, err := newYamlConfig([]byte(`element: value`))
	require.NoError(t, err, "Cannot parse yaml-config")

	for _, functors := range elementFunctors {
		_, err = functors.Getter(config, "")
		require.EqualError(t, err, ErrorNotFound.Error())
	}
}

func TestYamlGetAbsentValue(t *testing.T) {
	config, err := newYamlConfig([]byte(`element: value`))
	require.NoError(t, err, "Cannot parse yaml-config")

	for _, functors := range elementFunctors {
		_, err = functors.Getter(config, "/root")
		require.Error(t, err, ErrorNotFound.Error())
	}
}

func TestYamlGetAbsentConfigPart(t *testing.T) {
	config, err := newYamlConfig([]byte(`element: value`))
	require.NoError(t, err, "Cannot parse yaml-config")

	_, err = config.GetConfigPart("/root")
	require.Error(t, err, ErrorNotFound.Error())
}

func TestYamlGetValueOfIncorrectType(t *testing.T) {
	config, err := newYamlConfig([]byte(oneLevelYamlConfig))
	require.NoError(t, err, "Cannot parse xml-config")

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

func TestYamlGrabAbsentValue(t *testing.T) {
	config, err := newYamlConfig([]byte(oneLevelYamlConfig))
	require.NoError(t, err, "Cannot parse yaml-config")

	executed := false
	err = config.GrabValue("/absentElement", func(data interface{}) error {
		executed = true
		return nil
	})

	require.EqualError(t, err, ErrorNotFound.Error())
	require.Equal(t, false, executed, "Value grabber must not be executed")
}

func TestYamlGrabAbsentValues(t *testing.T) {
	config, err := newYamlConfig([]byte(oneLevelYamlConfig))
	require.NoError(t, err, "Cannot parse yaml-config")

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

func TestYamlGrabValuePassError(t *testing.T) {
	config, err := newYamlConfig([]byte(oneLevelYamlConfig))
	require.NoError(t, err, "Cannot parse yaml-config")

	expectedError := errors.New("TestYamlGrabValuePassError error")
	err = config.GrabValue("/intElement", func(data interface{}) error {
		return expectedError
	})

	require.EqualError(t, err, expectedError.Error())
}

func TestYamlGrabValuesPassError(t *testing.T) {
	config, err := newYamlConfig([]byte(oneLevelYamlConfig))
	require.NoError(t, err, "Cannot parse yaml-config")

	expectedError := errors.New("TestYamlGrabValuesPassError error")
	err = config.GrabValues("/intElements", DEFAULT_ARRAY_DELIMITER,
		func(length int) {},
		func(data interface{}) error {
			return expectedError
		})

	require.EqualError(t, err, expectedError.Error())
}

func TestYamlGrabValuesOfSingleElement(t *testing.T) {
	config, err := newYamlConfig([]byte(oneLevelYamlConfig))
	require.NoError(t, err, "Cannot parse yaml-config")

	err = config.GrabValues("/intElement", DEFAULT_ARRAY_DELIMITER,
		func(length int) {},
		func(data interface{}) error {
			return nil
		})

	require.EqualError(t, err, ErrorIncorrectValueType.Error())
}