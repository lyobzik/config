package config
import (
	"path"
	"strings"
	"errors"
	"io/ioutil"
	"io"
	"os"
)

const (
	CONF = "conf"
	INI = "ini"
	JSON = "json"
	XML = "xml"
	YAML = "yaml"
	YML = "yml"

	PATH_DELIMITER = "/"
)

var (
	ErrorNotFound = errors.New("Not found")
	ErrorIncorrectPath = errors.New("Incorrect path")
	ErrorUnknownConfigType = errors.New("Unknown config type")
)

type Config interface {
	GetValue(path string) (value interface{}, err error)

	GetString(path string) (value string, err error)
	GetBool(path string) (value bool, err error)
	GetFloat(path string) (value float64, err error)
	GetInt(path string) (value int64, err error)

	GetConfigPart(path string) (config Config, err error)

	LoadValue(path string, value interface{}) (err error)
}

func ReadConfig(configPath string) (Config, error) {
	return ReadTypedConfig(configPath, path.Ext(configPath))
}

func ReadTypedConfig(configPath string, configType string) (Config, error) {
	if len(configPath) == 0 {
		return nil, ErrorIncorrectPath
	}
	configFile, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	return ReadConfigFromReader(configFile, configType)
}

func ReadConfigFromReader(configReader io.Reader, configType string) (Config, error) {
	configData, err := ioutil.ReadAll(configReader)
	if err != nil {
		return nil, err
	}
	creator, err := getConfigCreator(configType)
	if err != nil {
		return nil, err
	}
	return creator(configData)
}

type configCreator func([]byte) (Config, error)

func getConfigCreator(configType string) (configCreator, error) {
	creators := map[string]configCreator{
		CONF: newIniConfig, INI: newIniConfig,
		JSON: newJsonConfig,
		XML: newXmlConfig,
		YAML: newYamlConfig, YML: newYamlConfig}

	if creator, exist := creators[configType]; exist {
		return creator, nil
	}
	return nil, ErrorUnknownConfigType
}

func splitPath(path string) ([]string) {
	path = strings.Trim(path, PATH_DELIMITER)
	if len(path) > 0 {
		return strings.Split(path, PATH_DELIMITER)
	}
	return []string{}
}