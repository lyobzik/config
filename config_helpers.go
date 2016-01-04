package config

import (
	"strings"
)

const (
	PATH_DELIMITER = "/"
)

// Heplers.
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
