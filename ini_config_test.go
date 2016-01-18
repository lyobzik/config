package config

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	oneLevelINIConfig = "stringElement=value\nboolElement=true\n" +
		"floatElement=1.23456\nintElement=123456\n" +
		"stringElements=value1 value2 value3\nboolElements=true false true\n" +
		"floatElements=1.23 4.56 7.89\nintElements=123 456 789\n" +
		"timeElement=2006-01-02T15:04:05+07:00\ndurationElement=2h45m5s150ms\n" +
		"timeElements=2006-01-02T15:04:05+07:00 2015-01-02T01:15:45Z 1999-12-31T23:59:59+00:00\n" +
		"durationElements=1h 1h15m30s450ms 1s750ms"

	twoLevelINIConfig = fmt.Sprintf("[first]\n%[1]s\n[second]\n%[1]s", oneLevelINIConfig)
)

func equalINITest(t *testing.T, data string, path string, functors Functors) {
	config, err := newINIConfig([]byte(data))
	require.Nil(t, err, "Cannot parse ini-config")

	value, err := functors.Getter(config, path)
	require.Nil(t, err, "Cannot get value of '%s'", path)

	functors.Checker(t, value)
}

// Tests.
func TestCreateEmptyIni(t *testing.T) {
	_, err := newINIConfig([]byte(""))
	require.Nil(t, err, "Cannot parse empty ini-config")
}

func TestOneLevelIni(t *testing.T) {
	for element, functors := range elementFunctors {
		equalINITest(t, oneLevelINIConfig, element, functors)
	}
}

func TestTwoLevelIni(t *testing.T) {
	for element, functors := range elementFunctors {
		equalINITest(t, twoLevelINIConfig, joinPath("first", element), functors)
		equalINITest(t, twoLevelINIConfig, joinPath("second", element), functors)
	}
}

func TestTwoLevelIniLoadValue(t *testing.T) {
	config, err := newINIConfig([]byte(twoLevelINIConfig))
	require.Nil(t, err, "Cannot parse ini-config")

	value := configData{}
	err = LoadValueIgnoringErrors(config, "/first", &value)
	require.Nil(t, err, "Cannot load value from config")

	value.Check(t)
}

func TestIniGetEmptyStrings(t *testing.T) {
	config, err := newINIConfig([]byte("element="))
	require.NoError(t, err, "Cannot parse ini-config")

	value, err := config.GetStrings("/element", defaultArrayDelimiter)
	require.NoError(t, err, "Cannot get value")

	require.Empty(t, value)
}

func TestIniGrabValue(t *testing.T) {
	config, err := newINIConfig([]byte(oneLevelINIConfig))
	require.NoError(t, err, "Cannot parse ini-config")

	var intValue int64
	var convertingError error
	var isStringData bool
	err = config.GrabValue("/intElement", func(data interface{}) error {
		var stringData string
		if stringData, isStringData = data.(string); !isStringData {
			return errors.New("Incorrect data type")
		}
		intValue, convertingError = parseINIInt(stringData)
		return nil
	})

	require.NoError(t, err, "Cannot grab value from ini-config")
	require.True(t, isStringData, "Data must be string")
	require.NoError(t, convertingError, "Cannot convert intElement to int")
	checkIntValue(t, intValue)
}

func TestIniGrabValues(t *testing.T) {
	config, err := newINIConfig([]byte(oneLevelINIConfig))
	require.NoError(t, err, "Cannot parse ini-config")

	var intValues []int64
	var isStringData bool
	err = config.GrabValues("/intElements", defaultArrayDelimiter,
		func(length int) { intValues = make([]int64, 0, length) },
		func(data interface{}) error {
			var stringData string
			if stringData, isStringData = data.(string); !isStringData {
				return errors.New("Incorrect data type")
			}
			value, err := parseINIInt(stringData)
			if err != nil {
				return err
			}
			intValues = append(intValues, value)
			return nil
		})

	require.NoError(t, err, "Cannot grab value from ini-config")
	require.True(t, isStringData, "Data must be string")
	checkIntValues(t, intValues)
}

func TestIniGrabValuesOfSingleElement(t *testing.T) {
	config, err := newINIConfig([]byte(oneLevelINIConfig))
	require.NoError(t, err, "Cannot parse ini-config")

	var intValues []int64
	var isStringData bool
	err = config.GrabValues("/intElement", defaultArrayDelimiter,
		func(length int) { intValues = make([]int64, 0, length) },
		func(data interface{}) error {
			var stringData string
			if stringData, isStringData = data.(string); !isStringData {
				return errors.New("Incorrect data type")
			}
			value, err := parseINIInt(stringData)
			if err != nil {
				return err
			}
			intValues = append(intValues, value)
			return nil
		})

	require.NoError(t, err, "Cannot grab value from ini-config")
	require.True(t, isStringData, "Data must be string")
	require.Len(t, intValues, 1)
	checkIntValue(t, intValues[0])
}

// Negative tests.
func TestIncorrectIniConfig(t *testing.T) {
	_, err := newINIConfig([]byte("{}"))
	require.Error(t, err, "Incorrect ini-config parsed successfully")
}

func TestIniGetValueEmptyPath(t *testing.T) {
	config, err := newINIConfig([]byte(""))
	require.NoError(t, err, "Cannot parse ini-config")

	for _, functors := range elementFunctors {
		_, err = functors.Getter(config, "")
		require.EqualError(t, err, ErrorNotFound.Error())
	}
}

func TestEmptyIniGetAbsentValue(t *testing.T) {
	config, err := newINIConfig([]byte(""))
	require.NoError(t, err, "Cannot parse ini-config")

	for _, functors := range elementFunctors {
		_, err = functors.Getter(config, "/")
		require.Error(t, err, "Parameter must be absent")

		_, err = functors.Getter(config, "/element")
		require.Error(t, err, "Parameter must be absent")

		_, err = functors.Getter(config, "/first/element")
		require.Error(t, err, "Parameter must be absent")

		_, err = functors.Getter(config, "/first/child/element")
		require.Error(t, err, "Parameter must be absent")
	}
}

func TestOneLevelIniGetAbsentValue(t *testing.T) {
	config, err := newINIConfig([]byte(oneLevelINIConfig))
	require.NoError(t, err, "Cannot parse ini-config")

	for _, functors := range elementFunctors {
		_, err = functors.Getter(config, "/")
		require.Error(t, err, "Parameter must be absent")

		_, err = functors.Getter(config, "/element")
		require.Error(t, err, "Parameter must be absent")

		_, err = functors.Getter(config, "/first/element")
		require.Error(t, err, "Parameter must be absent")

		_, err = functors.Getter(config, "/first/child/element")
		require.Error(t, err, "Parameter must be absent")
	}
}

func TestTwoLevelIniGetAbsentValue(t *testing.T) {
	config, err := newINIConfig([]byte(twoLevelINIConfig))
	require.NoError(t, err, "Cannot parse ini-config")

	for _, functors := range elementFunctors {
		_, err = functors.Getter(config, "/")
		require.Error(t, err, "Parameter must be absent")

		_, err = functors.Getter(config, "/element")
		require.Error(t, err, "Parameter must be absent")

		_, err = functors.Getter(config, "/first/element")
		require.Error(t, err, "Parameter must be absent")

		_, err = functors.Getter(config, "/first/child/element")
		require.Error(t, err, "Parameter must be absent")
	}
}

func TestIniGetValueOfIncorrectType(t *testing.T) {
	config, err := newINIConfig([]byte(oneLevelINIConfig))
	require.NoError(t, err, "Cannot parse ini-config")

	_, err = config.GetBool("/stringElement")
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetInt("/stringElement")
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetFloat("/stringElement")
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetBools("/stringElement", defaultArrayDelimiter)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetInts("/stringElement", defaultArrayDelimiter)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetFloats("/stringElement", defaultArrayDelimiter)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetBools("/stringElements", defaultArrayDelimiter)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetInts("/stringElements", defaultArrayDelimiter)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")

	_, err = config.GetFloats("/stringElements", defaultArrayDelimiter)
	require.Error(t, err, ErrorIncorrectValueType.Error(), "Incorrect value parsed successfully")
}

func TestIniGrabAbsentValue(t *testing.T) {
	config, err := newINIConfig([]byte(oneLevelINIConfig))
	require.NoError(t, err, "Cannot parse ini-config")

	executed := false
	err = config.GrabValue("/absentElement", func(data interface{}) error {
		executed = true
		return nil
	})

	require.EqualError(t, err, ErrorNotFound.Error())
	require.Equal(t, false, executed, "Value grabber must not be executed")
}

func TestIniGrabAbsentValues(t *testing.T) {
	config, err := newINIConfig([]byte(oneLevelINIConfig))
	require.NoError(t, err, "Cannot parse ini-config")

	executed := false
	err = config.GrabValues("/absentElement", defaultArrayDelimiter,
		func(length int) { executed = true },
		func(data interface{}) error {
			executed = true
			return nil
		})

	require.EqualError(t, err, ErrorNotFound.Error())
	require.Equal(t, false, executed, "Value grabber must not be executed")
}

func TestIniGrabValuePassError(t *testing.T) {
	config, err := newINIConfig([]byte(oneLevelINIConfig))
	require.NoError(t, err, "Cannot parse ini-config")

	expectedError := errors.New("TestIniGrabValuePassError error")
	err = config.GrabValue("/intElement", func(data interface{}) error {
		return expectedError
	})

	require.EqualError(t, err, expectedError.Error())
}

func TestIniGrabValuesPassError(t *testing.T) {
	config, err := newINIConfig([]byte(oneLevelINIConfig))
	require.NoError(t, err, "Cannot parse ini-config")

	expectedError := errors.New("TestIniGrabValuesPassError error")
	err = config.GrabValues("/intElements", defaultArrayDelimiter,
		func(length int) {},
		func(data interface{}) error {
			return expectedError
		})

	require.EqualError(t, err, expectedError.Error())
}

// Parser tests.
func TestParseIniBool(t *testing.T) {
	value, err := parseINIBool(fmt.Sprint(expectedBoolValue))
	require.NoError(t, err, "Cannot parse ini bool")
	checkBoolValue(t, value)

	_, err = parseINIBool(expectedStringValue)
	require.EqualError(t, err, ErrorIncorrectValueType.Error())
}

func TestParseIniFloat(t *testing.T) {
	value, err := parseINIFloat(fmt.Sprint(expectedFloatValue))
	require.NoError(t, err, "Cannot parse ini float")
	checkFloatValue(t, value)

	_, err = parseINIFloat(expectedStringValue)
	require.EqualError(t, err, ErrorIncorrectValueType.Error())
}

func TestParseIniInt(t *testing.T) {
	value, err := parseINIInt(fmt.Sprint(expectedIntValue))
	require.NoError(t, err, "Cannot parse ini int")
	checkIntValue(t, value)

	value, err = parseINIInt(fmt.Sprint(float64(expectedIntValue)))
	require.NoError(t, err, "Cannot parse ini int")
	checkIntValue(t, value)

	_, err = parseINIInt(fmt.Sprint(float64(expectedIntValue) + 0.00000001))
	require.EqualError(t, err, ErrorIncorrectValueType.Error())

	_, err = parseINIInt(expectedStringValue)
	require.EqualError(t, err, ErrorIncorrectValueType.Error())
}

// Test GetConfigPart
func TestIniGetConfigPartRootFromRoot(t *testing.T) {
	rootConfig, err := newINIConfig([]byte(twoLevelINIConfig))
	require.Nil(t, err, "Cannot parse root ini-config")

	configPart, err := rootConfig.GetConfigPart("/")
	require.Nil(t, err, "Cannot get config part")

	require.Equal(t, rootConfig, configPart, "Not equal configs")
}

func TestIniGetConfigPartSectionFromRoot(t *testing.T) {
	expectedConfig, err := newINIConfig([]byte(oneLevelINIConfig))
	require.Nil(t, err, "Cannot parse expected ini-config")

	expectedValue := configData{}
	err = LoadValueIgnoringErrors(expectedConfig, "/", &expectedValue)
	require.Nil(t, err, "Cannot load value from expected ini-config")

	rootConfig, err := newINIConfig([]byte(twoLevelINIConfig))
	require.Nil(t, err, "Cannot parse root ini-config")

	configPart, err := rootConfig.GetConfigPart("/first")
	require.Nil(t, err, "Cannot get config part")

	value := configData{}
	err = LoadValueIgnoringErrors(configPart, "/", &value)
	require.Nil(t, err, "Cannot load value from ini-config")

	require.Equal(t, value, expectedValue, "Not equal configs")
}

func TestIniGetConfigPartKeyFromRoot(t *testing.T) {
	rootConfig, err := newINIConfig([]byte(twoLevelINIConfig))
	require.NoError(t, err, "Cannot parse root ini-config")

	for element, functors := range elementFunctors {
		configPart, err := rootConfig.GetConfigPart(joinPath("/first", element))
		require.NoError(t, err, "Cannot get config part")

		value, err := functors.Getter(configPart, "/")
		require.NoError(t, err, "Cannot get value from config")

		functors.Checker(t, value)
	}
}

func TestIniGetConfigPartSectionFromSection(t *testing.T) {
	rootConfig, err := newINIConfig([]byte(twoLevelINIConfig))
	require.NoError(t, err, "Cannot parse root ini-config")

	configSection, err := rootConfig.GetConfigPart("/first")
	require.NoError(t, err, "Cannot get config section")

	configPart, err := rootConfig.GetConfigPart("/first")
	require.NoError(t, err, "Cannot get config section")

	require.Equal(t, configSection, configPart, "Not equal configs")
}

func TestIniGetConfigPartKeyFromSection(t *testing.T) {
	rootConfig, err := newINIConfig([]byte(twoLevelINIConfig))
	require.NoError(t, err, "Cannot parse root ini-config")

	configSection, err := rootConfig.GetConfigPart("/first")
	require.NoError(t, err, "Cannot get config section")

	for element, functors := range elementFunctors {
		configPart, err := configSection.GetConfigPart(element)
		require.NoError(t, err, "Cannot get config part")

		value, err := functors.Getter(configPart, "/")
		require.NoError(t, err, "Cannot get value from config")

		functors.Checker(t, value)
	}
}

func TestIniGetConfigPartKeyFromKey(t *testing.T) {
	rootConfig, err := newINIConfig([]byte(twoLevelINIConfig))
	require.NoError(t, err, "Cannot parse root ini-config")

	for element, functors := range elementFunctors {
		configKey, err := rootConfig.GetConfigPart(joinPath("first", element))
		require.NoError(t, err, "Cannot get config key")

		configPart, err := configKey.GetConfigPart("/")
		require.NoError(t, err, "Cannot get config part")

		value, err := functors.Getter(configPart, "/")
		require.NoError(t, err, "Cannot get value from config")

		functors.Checker(t, value)
	}
}

func TestIniGetConfigPartWithLongPath(t *testing.T) {
	rootConfig, err := newINIConfig([]byte(twoLevelINIConfig))
	require.NoError(t, err, "Cannot parse root ini-config")

	configSection, err := rootConfig.GetConfigPart("/first")
	require.NoError(t, err, "Cannot get config section")

	configKey, err := rootConfig.GetConfigPart("/first/stringElement")
	require.NoError(t, err, "Cannot get config key ection")

	_, err = rootConfig.GetConfigPart("/first/stringElement/element")
	require.EqualError(t, err, ErrorNotFound.Error())

	_, err = configSection.GetConfigPart("/stringElement/element")
	require.EqualError(t, err, ErrorNotFound.Error())

	_, err = configKey.GetConfigPart("/element")
	require.EqualError(t, err, ErrorNotFound.Error())
}

func TestIniGetConfigPartAbsentSectionFromTwoLevelRoot(t *testing.T) {
	rootConfig, err := newINIConfig([]byte(twoLevelINIConfig))
	require.NoError(t, err, "Cannot parse root ini-config")

	_, err = rootConfig.GetConfigPart("/third")
	require.EqualError(t, err, ErrorNotFound.Error())
}

func TestIniGetConfigPartAbsentKeyFromTwoLevelRoot(t *testing.T) {
	rootConfig, err := newINIConfig([]byte(twoLevelINIConfig))
	require.NoError(t, err, "Cannot parse root ini-config")

	_, err = rootConfig.GetConfigPart("/first/element")
	require.EqualError(t, err, ErrorNotFound.Error())

	_, err = rootConfig.GetConfigPart("/third/stringElement")
	require.EqualError(t, err, ErrorNotFound.Error())
}

func TestIniGetConfigPartAbsentKeyFromSection(t *testing.T) {
	rootConfig, err := newINIConfig([]byte(twoLevelINIConfig))
	require.NoError(t, err, "Cannot parse root ini-config")

	configSection, err := rootConfig.GetConfigPart("/first")
	require.NoError(t, err, "Cannot get config section")

	_, err = configSection.GetConfigPart("element")
	require.EqualError(t, err, ErrorNotFound.Error())
}
