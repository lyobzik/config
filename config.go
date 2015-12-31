package config
import (
	"path"
	"strings"
	"errors"
)

const (
	INI = "ini"
	JSON = "json"
	XML = "xml"
	YAML = "yaml"

	PATH_DELIMITER = "/"
)

var (
	ErrorNotFound = errors.New("Not found")
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

func ReadConfig(configPath string) (config *Config, err error) {
	return ReadTypedConfig(configPath, path.Ext(configPath))
}

func ReadTypedConfig(configPath string, configType string) (config *Config, err error) {
	return nil, nil
}

func splitPath(path string) ([]string) {
	path = strings.Trim(path, PATH_DELIMITER)
	return strings.Split(path, PATH_DELIMITER)
}