package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	errorIncorrectConfigReader = errors.New("Incorrect config reader")
)

// Tests.
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

// Negative tests.
func TestEmptyPathToConfig(t *testing.T) {
	_, err := ReadConfig("")
	require.EqualError(t, err, ErrorIncorrectPath.Error())
}

func TestIncorrectPathToConfig(t *testing.T) {
	_, err := ReadConfig("/")
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

	var initValue int = 5
	var value int = initValue
	err = LoadValue(config, "/", &value)
	require.EqualError(t, err, ErrorNotFound.Error())
	require.Equal(t, initValue, value, "Value must be unchanged")

	err = LoadValueIgnoringErrors(config, "/", &value)
	require.NoError(t, err, "Cannot load value from config")
	require.Equal(t, initValue, value, "Value must be unchanged")
}
