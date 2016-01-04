package config

import (
	"errors"

	ini "gopkg.in/ini.v1"
	"strconv"
)

type iniConfig struct {
	file *ini.File
	section *ini.Section
	key *ini.Key
}

func newIniConfig(data []byte) (Config, error) {
	file, err := ini.Load(data)
	if err != nil {
		return nil, err
	}

	return &iniConfig{file: file}, nil
}

func (c *iniConfig) GetValue(path string) (value interface{}, err error) {
	configPart, err := c.GetConfigPart(path)
	if err != nil {
		return nil, err
	}
	iniConfigPart, converted := configPart.(*iniConfig)
	if converted && iniConfigPart.key != nil {
		return iniConfigPart.key.Value(), nil
	}

	return nil, errors.New("Not implemented")
}

func (c *iniConfig) GetString(path string) (value string, err error) {
	key, err := c.FindKey(path)
	if err != nil {
		return value, err
	}
	return key.String(), nil
}

func (c *iniConfig) GetBool(path string) (value bool, err error) {
	key, err := c.FindKey(path)
	if err != nil {
		return value, err
	}
	return key.Bool()
}

func (c *iniConfig) GetFloat(path string) (value float64, err error) {
	key, err := c.FindKey(path)
	if err != nil {
		return value, err
	}
	return key.Float64()
}

func (c *iniConfig) GetInt(path string) (value int64, err error) {
	key, err := c.FindKey(path)
	if err != nil {
		return value, err
	}
	return key.Int64()
}

func (c *iniConfig) GetStrings(path string, delim string) (value []string, err error) {
	key, err := c.FindKey(path)
	if err != nil {
		return value, err
	}
	return key.Strings(delim), nil
}

func (c *iniConfig) GetBools(path string, delim string) (value []bool, err error) {
	key, err := c.FindKey(path)
	if err != nil {
		return value, err
	}
	stringValues := key.Strings(delim)
	value = make([]bool, len(stringValues))
	for i := range stringValues {
		value[i], err = strconv.ParseBool(stringValues[i])
		if err != nil {
			return value, err
		}
	}
	return value, nil
}

func (c *iniConfig) GetFloats(path string, delim string) (value []float64, err error) {
	key, err := c.FindKey(path)
	if err != nil {
		return value, err
	}
	return key.Float64s(delim), nil
}

func (c *iniConfig) GetInts(path string, delim string) (value []int64, err error) {
	key, err := c.FindKey(path)
	if err != nil {
		return value, err
	}
	return key.Int64s(delim), nil
}

func (c *iniConfig) GetConfigPart(path string) (Config, error) {
	pathParts := splitPath(path)
	if len(pathParts) == 0 {
		return c, nil
	}

	if c.key != nil {
		if len(pathParts) > 0 {
			return nil, ErrorIncorrectPath
		}
		return &iniConfig{key: c.key}, nil
	}

	if c.section != nil {
		if len(pathParts) > 1 {
			return nil, ErrorIncorrectPath
		}
		if len(pathParts) == 0 {
			return &iniConfig{section: c.section}, nil
		}
		key := c.section.Key(pathParts[0])
		if key == nil {
			return nil, ErrorNotFound
		}
		return &iniConfig{key: key}, nil
	}

	// else c.file != nil
	if len(pathParts) > 2 {
		return nil, ErrorIncorrectPath
	}
	if len(pathParts) == 0 {
		return &iniConfig{file: c.file}, nil
	}
	var section *ini.Section
	if len(pathParts) == 1 && len(c.file.Sections()) == 1 {
		section = c.file.Section("")
	} else {
		section = c.file.Section(pathParts[0])
	}
	if section == nil {
		return nil, ErrorNotFound
	}
	if len(pathParts) == 1 {
		return &iniConfig{section: section}, nil
	}
	key := section.Key(pathParts[len(pathParts) - 1])
	if key == nil {
		return nil, ErrorNotFound
	}
	return &iniConfig{key: key}, nil
}

func (c *iniConfig) LoadValue(path string, value interface{}) (err error) {
	configPart, err := c.GetConfigPart(path)
	if err != nil {
		return err
	}
	iniConfigPart := configPart.(*iniConfig)
	if iniConfigPart.key != nil {
		return errors.New("Not implemented")
	}

	if iniConfigPart.section != nil {
		return iniConfigPart.section.MapTo(value)
	}

	return iniConfigPart.file.MapTo(value)
}

func (c *iniConfig) FindKey(path string) (*ini.Key, error) {
	pathParts := splitPath(path)
	if c.key != nil {
		if len(pathParts) != 0 {
			return nil, ErrorIncorrectPath
		}
		return c.key, nil
	}

	if c.section != nil {
		if len(pathParts) != 1 {
			return nil, ErrorIncorrectPath
		}
		key := c.section.Key(pathParts[0])
		if key == nil {
			return nil, ErrorNotFound
		}
		return key, nil
	}

	if c.file != nil {
		if len(pathParts) < 1 || 2 < len(pathParts) {
			return nil, ErrorIncorrectPath
		}
		var section *ini.Section
		if len(pathParts) == 1 && len(c.file.Sections()) == 1 {
			section = c.file.Section("")
		} else {
			section = c.file.Section(pathParts[0])
		}
		if section == nil {
			return nil, ErrorNotFound
		}
		key := section.Key(pathParts[len(pathParts) - 1])
		if key == nil {
			return nil, ErrorNotFound
		}
		return key, nil
	}
	return nil, ErrorNotFound
}
