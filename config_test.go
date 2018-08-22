//Copyright 2016 lyobzik
//
//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package config

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var (
	errorIncorrectConfigReader = errors.New("Incorrect config reader")
)

// Tests.
func TestGetConfigType(t *testing.T) {
	for _, expectedType := range []string{CONF, INI, JSON, XML, YAML, YML} {
		configType1 := getConfigType("/etc/config." + expectedType)
		require.Equal(t, expectedType, configType1)
		configType2 := getConfigType("/etc/c.o.n.f.i.g." + expectedType)
		require.Equal(t, expectedType, configType2)
		configType3 := getConfigType("/etc/." + expectedType)
		require.Equal(t, expectedType, configType3)
		configType4 := getConfigType("config." + expectedType)
		require.Equal(t, expectedType, configType4)
		configType5 := getConfigType("." + expectedType)
		require.Equal(t, expectedType, configType5)
	}

	for _, expectedType := range []string{CONF, INI, JSON, XML, YAML, YML} {
		configType1 := getConfigType("/etc/" + expectedType)
		require.Equal(t, "", configType1)
		configType2 := getConfigType("/etc/.")
		require.Equal(t, "", configType2)
		configType3 := getConfigType(expectedType)
		require.Equal(t, "", configType3)
		configType4 := getConfigType(".")
		require.Equal(t, "", configType4)
		configType5 := getConfigType("")
		require.Equal(t, "", configType5)
	}
}

func TestCreatedConfigTypes(t *testing.T) {
	conf, err := CreateConfigFromString("", CONF)
	require.NoError(t, err, "Cannot create conf-config")
	require.IsType(t, (*iniConfig)(nil), conf, "Incorrect type of created config")

	ini, err := CreateConfigFromString("", INI)
	require.NoError(t, err, "Cannot create ini-config")
	require.IsType(t, (*iniConfig)(nil), ini, "Incorrect type of created config")

	json, err := CreateConfigFromString("{}", JSON)
	require.NoError(t, err, "Cannot create conf-config")
	require.IsType(t, (*jsonConfig)(nil), json, "Incorrect type of created config")

	xml, err := CreateConfigFromString("", XML)
	require.NoError(t, err, "Cannot create conf-config")
	require.IsType(t, (*xmlConfig)(nil), xml, "Incorrect type of created config")

	yaml, err := CreateConfigFromString("", YAML)
	require.NoError(t, err, "Cannot create conf-config")
	require.IsType(t, (*yamlConfig)(nil), yaml, "Incorrect type of created config")

	yml, err := CreateConfigFromString("", YML)
	require.NoError(t, err, "Cannot create conf-config")
	require.IsType(t, (*yamlConfig)(nil), yml, "Incorrect type of created config")
}

func TestCreateConfigFromReader(t *testing.T) {
	reader := bytes.NewReader([]byte(oneLevelJSONConfig))
	_, err := ReadConfigFromReader(reader, JSON)
	require.NoError(t, err, "Cannot read config")
}

func TestTimeDefaultLoader(t *testing.T) {
	loader, exist := defaultLoaders["time.Time"]
	require.True(t, exist, "Cannot found loader for time.Time")

	loadedValue, err := loader(expectedTimeValue.Format(time.RFC3339))
	require.NoError(t, err, "Cannot load time value")

	var value time.Time
	reflect.ValueOf(&value).Elem().Set(loadedValue)
	require.Equal(t, expectedTimeValue, value)

	_, err = loader("time value")
	require.Error(t, err)
}

func TestDurationDefaultLoader(t *testing.T) {
	loader, exist := defaultLoaders["time.Duration"]
	require.True(t, exist, "Cannot found loader for time.Duration")

	loadedValue, err := loader(expectedDurationValue.String())
	require.NoError(t, err, "Cannot load duration value")

	var value time.Duration
	reflect.ValueOf(&value).Elem().Set(loadedValue)
	require.Equal(t, expectedDurationValue, value)

	_, err = loader("duration value")
	require.Error(t, err)
}

// Negative tests.
func TestEmptyPathToConfig(t *testing.T) {
	_, err := ReadConfig("")
	require.EqualError(t, err, ErrorIncorrectPath.Error())
}

func TestIncorrectPathToConfig(t *testing.T) {
	_, err := ReadConfig("/incorrectConfigPath")
	require.Error(t, err)
}

func TestReadExistButIncorrectFile(t *testing.T) {
	_, err := ReadConfig(".")
	require.Error(t, err)
}

type IncorrectConfigReader struct {
}

func (c IncorrectConfigReader) Read([]byte) (int, error) {
	return 0, errors.New("Incorrect config reader")
}

func TestIncorrectCofigReader(t *testing.T) {
	reader := IncorrectConfigReader{}
	_, err := ReadConfigFromReader(reader, CONF)
	require.EqualError(t, err, errorIncorrectConfigReader.Error())
}

func TestIncorrectConfigType(t *testing.T) {
	_, err := CreateConfig([]byte{}, "unknownType")
	require.EqualError(t, err, ErrorUnknownConfigType.Error())
}

// LoadValue tests.
func TestLoadEmptyConfig(t *testing.T) {
	config, err := CreateConfigFromString("{}", JSON)
	require.NoError(t, err, "Cannot load config")

	initValue := 5
	value := initValue
	err = LoadValue(config, "/", &value)
	require.EqualError(t, err, ErrorNotFound.Error())
	require.Equal(t, initValue, value, "Value must be unchanged")

	err = LoadValueIgnoringMissingFieldErrors(config, "/", &value)
	require.NoError(t, err, "Cannot load value from config")
	require.Equal(t, initValue, value, "Value must be unchanged")
}

type StructWithUnsignedIntegerField struct {
	UnsignedElement uint64 `config:"unsignedElement"`
}

func TestLoadValueWithUnsignedIntegerField(t *testing.T) {
	config, err := CreateConfigFromString(`{"unsignedElement": 123456}`, JSON)
	require.NoError(t, err, "Cannot load config")

	var value StructWithUnsignedIntegerField
	err = LoadValue(config, "/", &value)
	require.NoError(t, err, "Cannot load value with unsigned field")
	checkEqual(t, value.UnsignedElement, uint64(123456))
}

type StructWithUnsignedIntegerSlice struct {
	UnsignedElements []uint64 `config:"unsignedElements"`
}

func TestLoadValueWithUnsignedIntegerSlice(t *testing.T) {
	config, err := CreateConfigFromString(`{"unsignedElements": [123, 456, 789]}`, JSON)
	require.NoError(t, err, "Cannot load config")

	var value StructWithUnsignedIntegerSlice
	err = LoadValue(config, "/", &value)
	require.NoError(t, err, "Cannot load value with unsigned slice")
	checkEqual(t, value.UnsignedElements, []uint64{123, 456, 789})
}

type StructWithPtrToUnsignedIntegerField struct {
	UnsignedElement *uint64 `config:"unsignedElement"`
}

func TestLoadValueWithPtrToUnsignedIntegerField(t *testing.T) {
	config, err := CreateConfigFromString(`{"unsignedElement": 123456}`, JSON)
	require.NoError(t, err, "Cannot load config")

	var value StructWithPtrToUnsignedIntegerField
	err = LoadValue(config, "/", &value)
	require.NoError(t, err, "Cannot load value with unsigned field")
	checkEqual(t, *value.UnsignedElement, uint64(123456))
}

func TestLoadValueWithMissingPtrToUnsignedIntegerField(t *testing.T) {
	config, err := CreateConfigFromString(`{"notUnsignedElement": 123456}`, JSON)
	require.NoError(t, err, "Cannot load config")

	var unsignedElement uint64 = 1234567
	value := StructWithPtrToUnsignedIntegerField{UnsignedElement: &unsignedElement}
	err = LoadValueIgnoringMissingFieldErrors(config, "/", &value)
	require.NoError(t, err, "Cannot load value with unsigned field")
	checkEqual(t, value.UnsignedElement, &unsignedElement)
	checkEqual(t, *value.UnsignedElement, unsignedElement)
}

func TestLoadValueWithPtrToUnsignedIntegerFieldFromInvalidValueType(t *testing.T) {
	config, err := CreateConfigFromString(`{"unsignedElement": "invalid type"}`, JSON)
	require.NoError(t, err, "Cannot load config")

	var value StructWithPtrToUnsignedIntegerField
	err = LoadValue(config, "/", &value)
	require.EqualError(t, err, ErrorIncorrectValueType.Error())
}

type StructWithPtrToStructField struct {
	StructElement *StructWithPtrToUnsignedIntegerField `config:"structElement"`
}

func TestLoadValueWithPtrToStructField(t *testing.T) {
	config, err := CreateConfigFromString(`{"structElement": {"unsignedElement": 123456}}`, JSON)
	require.NoError(t, err, "Cannot load config")

	var value StructWithPtrToStructField
	err = LoadValue(config, "/", &value)
	require.NoError(t, err, "Cannot load value with unsigned field")
	checkEqual(t, *value.StructElement.UnsignedElement, uint64(123456))
}

type LoadableStruct struct {
	IntElement   int64
	FloatElement float64
}

var (
	errorForTestLoadLoadableValue = errors.New("Cannot load loadable value")
)

func (s *LoadableStruct) LoadValueFromConfig(data string) (err error) {
	values := strings.Split(data, " ")
	if len(values) != 2 {
		return errorForTestLoadLoadableValue
	}
	if s.IntElement, err = strconv.ParseInt(values[0], 10, 64); err != nil {
		return errorForTestLoadLoadableValue
	}
	if s.FloatElement, err = strconv.ParseFloat(values[1], 64); err != nil {
		return errorForTestLoadLoadableValue
	}
	return nil
}

type StructWithLoadableField struct {
	Value  LoadableStruct
	Values []LoadableStruct
}

func TestLoadValueWithLoadableField(t *testing.T) {
	config, err := CreateConfigFromString(`{"Value": "123456 1.23456",
		"Values": ["123 1.23", "456 4.56", "789 7.89"]}`, JSON)
	require.NoError(t, err, "Cannot load config")

	var value StructWithLoadableField
	err = LoadValue(config, "/", &value)
	require.NoError(t, err, "Cannot load value with loadable field")
	checkEqual(t, value.Value, LoadableStruct{IntElement: expectedIntValue,
		FloatElement: expectedFloatValue})
	checkEqual(t, value.Values, []LoadableStruct{
		{IntElement: expectedIntValues[0], FloatElement: expectedFloatValues[0]},
		{IntElement: expectedIntValues[1], FloatElement: expectedFloatValues[1]},
		{IntElement: expectedIntValues[2], FloatElement: expectedFloatValues[2]}})
}

type UnloadableStruct struct {
	IntElement   int64
	FloatElement float64
}

type StructWithUnloadableField struct {
	Value UnloadableStruct
}

func TestLoadValueWithUnloadableField(t *testing.T) {
	config, err := CreateConfigFromString(
		`{"Value": {"IntElement": 123456, "FloatElement": 1.23456}}`, JSON)
	require.NoError(t, err, "Cannot load config")

	var value StructWithUnloadableField
	err = LoadValue(config, "/", &value)
	require.NoError(t, err, "Cannot load value with unloadable field")
	checkEqual(t, value.Value, UnloadableStruct{IntElement: expectedIntValue,
		FloatElement: expectedFloatValue})
}

func TestLoadValueToIncorrectVariable(t *testing.T) {
	config, err := CreateConfigFromString("{}", JSON)
	require.NoError(t, err, "Cannot load config")

	var value int
	err = LoadValue(config, "/", value)
	require.EqualError(t, err, ErrorIncorrectValueToLoadFromConfig.Error())
}

type StructWithIncorrectFieldType struct {
	Value interface{}
}

func TestLoadValueWithIncorrectFieldType(t *testing.T) {
	config, err := CreateConfigFromString("{}", JSON)
	require.NoError(t, err, "Cannot load config")

	var value StructWithIncorrectFieldType
	err = LoadValue(config, "/", &value)
	require.EqualError(t, err, ErrorUnsupportedTypeToLoadValue.Error())
}

func TestLoadValueWithIncorrectFieldTypeIgnoringErrors(t *testing.T) {
	config, err := CreateConfigFromString("{}", JSON)
	require.NoError(t, err, "Cannot load config")

	var value StructWithIncorrectFieldType

	settings := GetDefaultLoadSettings(true)
	settings.IgnoreUnsupportedTypeErrors = true
	err = TunedLoadValue(config, settings, "/", &value)
	require.NoError(t, err, "Cannot load config with unsupported fields")
}

type StructWithIncorrectSliceElementType struct {
	Value []StructWithIncorrectFieldType
}

func TestLoadValueWithIncorrectSliceElementType(t *testing.T) {
	config, err := CreateConfigFromString("{}", JSON)
	require.NoError(t, err, "Cannot load config")

	var value StructWithIncorrectSliceElementType
	err = LoadValue(config, "/", &value)
	require.EqualError(t, err, ErrorUnsupportedTypeToLoadValue.Error())
}

func TestLoadValueSliceWithError(t *testing.T) {
	config, err := CreateConfigFromString(`{"Value": "123456 1.23456",
		"Values": ["123 1.23", "456", "789"]}`, JSON)
	require.NoError(t, err, "Cannot load config")

	var value StructWithLoadableField
	err = LoadValue(config, "/", &value)

	require.EqualError(t, err, errorForTestLoadLoadableValue.Error())
	checkEqual(t, value.Value, LoadableStruct{IntElement: expectedIntValue,
		FloatElement: expectedFloatValue})
}

func TestLoadValueWithLoadableFieldIncorrectType(t *testing.T) {
	config, err := CreateConfigFromString(`{"Value": 123456, "Values": [123, 456, 789]}`, JSON)
	require.NoError(t, err, "Cannot load config")

	var value StructWithLoadableField
	err = LoadValue(config, "/", &value)
	require.EqualError(t, err, errorForTestLoadLoadableValue.Error())
}

func TestLoadValueWithLoadableFieldLoadError(t *testing.T) {
	config, err := CreateConfigFromString(`{"Value": "123456", "Values": ["123", "4.56", "789"]}`, JSON)
	require.NoError(t, err, "Cannot load config")

	var value StructWithLoadableField
	err = LoadValue(config, "/", &value)
	require.EqualError(t, err, errorForTestLoadLoadableValue.Error())
}

type StructWithDurationField struct {
	Value time.Duration `config:"nonexist_path"`
}

func TestLoadValueWithDurationField(t *testing.T) {
	config, err := CreateConfigFromString(`{"Value": "1ms"}`, JSON)
	require.NoError(t, err, "Cannot load config")

	var value StructWithDurationField
	err = LoadValue(config, "/", &value)
	require.EqualError(t, err, ErrorNotFound.Error())
}

type StructWithDurationsField struct {
	Values []time.Duration
}

func TestLoadValueWithDurationsField(t *testing.T) {
	config, err := CreateConfigFromString(`{"Values": [{}, {}, {}]}`, JSON)
	require.NoError(t, err, "Cannot load config")

	var value StructWithDurationsField
	err = LoadValue(config, "/", &value)
	require.EqualError(t, err, ErrorIncorrectValueType.Error())
}

// Test loading numeric types.
func loadValidValueFromConfig(t *testing.T, configContent, path string, value interface{}) {
	config, err := CreateConfigFromString(configContent, JSON)
	require.NoError(t, err, "Cannot load config")

	err = LoadValue(config, "/", value)
	require.NoError(t, err, "Cannot load value from config")
}

func loadValidValue(t *testing.T, expectedValue interface{}, value interface{}) {
	config := fmt.Sprintf(`{"Value": %v}`, expectedValue)
	loadValidValueFromConfig(t, config, "/", value)
}

func TestLoadIntValue(t *testing.T) {
	var value struct{ Value int }
	loadValidValue(t, expectedIntValue, &value)
	checkIntValue(t, int64(value.Value))
}

func TestLoadInt8Value(t *testing.T) {
	var value struct{ Value int8 }
	loadValidValue(t, expectedInt8Value, &value)
	checkEqual(t, value.Value, expectedInt8Value)
}

func TestLoadInt16Value(t *testing.T) {
	var value struct{ Value int16 }
	loadValidValue(t, expectedInt16Value, &value)
	checkEqual(t, value.Value, expectedInt16Value)
}

func TestLoadInt32Value(t *testing.T) {
	var value struct{ Value int32 }
	loadValidValue(t, expectedIntValue, &value)
	checkIntValue(t, int64(value.Value))
}

func TestLoadInt64Value(t *testing.T) {
	var value struct{ Value int64 }
	loadValidValue(t, expectedIntValue, &value)
	checkIntValue(t, value.Value)
}

func TestLoadUIntValue(t *testing.T) {
	var value struct{ Value int }
	loadValidValue(t, expectedIntValue, &value)
	checkIntValue(t, int64(value.Value))
}

func TestLoadUInt8Value(t *testing.T) {
	var value struct{ Value uint8 }
	loadValidValue(t, expectedInt8Value, &value)
	checkEqual(t, value.Value, uint8(expectedInt8Value))
}

func TestLoadUInt16Value(t *testing.T) {
	var value struct{ Value uint16 }
	loadValidValue(t, expectedInt16Value, &value)
	checkEqual(t, value.Value, uint16(expectedInt16Value))
}

func TestLoadUInt32Value(t *testing.T) {
	var value struct{ Value uint32 }
	loadValidValue(t, expectedIntValue, &value)
	checkIntValue(t, int64(value.Value))
}

func TestLoadUInt64Value(t *testing.T) {
	var value struct{ Value uint64 }
	loadValidValue(t, expectedIntValue, &value)
	checkIntValue(t, int64(value.Value))
}

func TestLoadFloat32Value(t *testing.T) {
	var value struct{ Value float32 }
	loadValidValue(t, expectedFloatValue, &value)
	checkFloatValue(t, float64(value.Value))
}

func TestLoadFloat64Value(t *testing.T) {
	var value struct{ Value float64 }
	loadValidValue(t, expectedFloatValue, &value)
	checkFloatValue(t, value.Value)
}

// Test loading slice of numeric types.
func loadValidValues(t *testing.T, value interface{}, expectedValues reflect.Value) {
	stringExpectedValues := make([]string, expectedValues.Len())
	for i := 0; i < expectedValues.Len(); i += 1 {
		stringExpectedValues[i] = fmt.Sprintf("%v", expectedValues.Index(i).Interface())
	}
	config := fmt.Sprintf(`{"Values": [%s]}`, strings.Join(stringExpectedValues, ", "))
	loadValidValueFromConfig(t, config, "/", value)
}

func convertToInt64Slice(values reflect.Value) []int64 {
	result := make([]int64, values.Len())
	for i := 0; i < values.Len(); i += 1 {
		result[i] = values.Index(i).Convert(reflect.TypeOf(int64(0))).Int()
	}
	return result
}

func TestLoadIntValues(t *testing.T) {
	var value struct{ Values []int }
	loadValidValues(t, &value, reflect.ValueOf(expectedIntValues))
	values := convertToInt64Slice(reflect.ValueOf(value.Values))
	checkIntValues(t, values)
}

func TestLoadInt8Values(t *testing.T) {
	var value struct{ Values []int8 }
	loadValidValues(t, &value, reflect.ValueOf(expectedInt8Values))
	checkEqual(t, value.Values, expectedInt8Values)
}

func TestLoadInt16Values(t *testing.T) {
	var value struct{ Values []int16 }
	loadValidValues(t, &value, reflect.ValueOf(expectedIntValues))
	values := convertToInt64Slice(reflect.ValueOf(value.Values))
	checkIntValues(t, values)
}

func TestLoadInt32Values(t *testing.T) {
	var value struct{ Values []int32 }
	loadValidValues(t, &value, reflect.ValueOf(expectedIntValues))
	values := convertToInt64Slice(reflect.ValueOf(value.Values))
	checkIntValues(t, values)
}

func TestLoadInt64Values(t *testing.T) {
	var value struct{ Values []int64 }
	loadValidValues(t, &value, reflect.ValueOf(expectedIntValues))
	values := convertToInt64Slice(reflect.ValueOf(value.Values))
	checkIntValues(t, values)
}

func TestLoadUIntValues(t *testing.T) {
	var value struct{ Values []uint }
	loadValidValues(t, &value, reflect.ValueOf(expectedIntValues))
	values := convertToInt64Slice(reflect.ValueOf(value.Values))
	checkIntValues(t, values)
}

func TestLoadUInt8Values(t *testing.T) {
	var value struct{ Values []uint8 }
	loadValidValues(t, &value, reflect.ValueOf(expectedInt8Values))
	values := convertToInt64Slice(reflect.ValueOf(value.Values))
	expectedValues := convertToInt64Slice(reflect.ValueOf(expectedInt8Values))
	checkEqual(t, values, expectedValues)
}

func TestLoadUInt16Values(t *testing.T) {
	var value struct{ Values []uint16 }
	loadValidValues(t, &value, reflect.ValueOf(expectedIntValues))
	values := convertToInt64Slice(reflect.ValueOf(value.Values))
	checkIntValues(t, values)
}

func TestLoadUInt32Values(t *testing.T) {
	var value struct{ Values []uint32 }
	loadValidValues(t, &value, reflect.ValueOf(expectedIntValues))
	values := convertToInt64Slice(reflect.ValueOf(value.Values))
	checkIntValues(t, values)
}

func TestLoadUInt64Values(t *testing.T) {
	var value struct{ Values []uint64 }
	loadValidValues(t, &value, reflect.ValueOf(expectedIntValues))
	values := convertToInt64Slice(reflect.ValueOf(value.Values))
	checkIntValues(t, values)
}

func TestLoadFloat32Values(t *testing.T) {
	var value struct{ Values []float32 }
	loadValidValues(t, &value, reflect.ValueOf(expectedFloatValues))

	values := make([]float64, 0)
	for _, value := range value.Values {
		values = append(values, float64(value))
	}
	checkFloatValues(t, values)
}

func TestLoadFloat64Values(t *testing.T) {
	var value struct{ Values []float64 }
	loadValidValues(t, &value, reflect.ValueOf(expectedFloatValues))
	checkFloatValues(t, value.Values)
}
