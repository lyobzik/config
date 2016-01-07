package config

import (
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

// Grabbers.
func (c *iniConfig) GrabValue(path string, grabber ValueGrabber) (err error) {
	value, err := c.GetString(path)
	if err != nil {
		return err
	}
	return grabber(value)
}

func (c *iniConfig) GrabValues(path string, delim string,
	creator ValueSliceCreator, grabber ValueGrabber) (err error) {

	values, err := c.GetStrings(path, delim)
	if err != nil {
		return err
	}
	creator(len(values))
	for _, value := range values {
		if err = grabber(value); err != nil {
			return err
		}
	}
	return nil
}

// Get single value.
func (c *iniConfig) GetString(path string) (value string, err error) {
	key, err := c.findKey(path)
	if err != nil {
		return value, err
	}
	return key.String(), nil
}

func (c *iniConfig) GetBool(path string) (value bool, err error) {
	key, err := c.findKey(path)
	if err != nil {
		return value, err
	}
	return key.Bool()
}

func (c *iniConfig) GetFloat(path string) (value float64, err error) {
	key, err := c.findKey(path)
	if err != nil {
		return value, err
	}
	return key.Float64()
}

func (c *iniConfig) GetInt(path string) (value int64, err error) {
	key, err := c.findKey(path)
	if err != nil {
		return value, err
	}
	return key.Int64()
}

// Get array of values.
func (c *iniConfig) GetStrings(path string, delim string) (value []string, err error) {
	key, err := c.findKey(path)
	if err != nil {
		return value, err
	}
	return key.Strings(delim), nil
}

func (c *iniConfig) GetBools(path string, delim string) (value []bool, err error) {
	return value, GrabStringValues(c, path, delim,
		func(cap int) { value = make([]bool, 0, cap) },
		func(data string) error {
			var parsed bool
			if parsed, err = strconv.ParseBool(data); err == nil {
				value = append(value, parsed)
			}
			return err
		})
}

func (c *iniConfig) GetFloats(path string, delim string) (value []float64, err error) {
	key, err := c.findKey(path)
	if err != nil {
		return value, err
	}
	return key.Float64s(delim), nil
}

func (c *iniConfig) GetInts(path string, delim string) (value []int64, err error) {
	key, err := c.findKey(path)
	if err != nil {
		return value, err
	}
	return key.Int64s(delim), nil
}

// Get subconfig.
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
		key, _ := c.section.GetKey(pathParts[0])
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
		section, _ = c.file.GetSection("")
	} else {
		section, _ = c.file.GetSection(pathParts[0])
	}
	if section == nil {
		return nil, ErrorNotFound
	}
	if len(pathParts) == 1 {
		return &iniConfig{section: section}, nil
	}
	key, _ := section.GetKey(pathParts[len(pathParts) - 1])
	if key == nil {
		return nil, ErrorNotFound
	}
	return &iniConfig{key: key}, nil
}

// Ini helpers.
func (c *iniConfig) findKey(path string) (*ini.Key, error) {
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
		key, _ := c.section.GetKey(pathParts[0])
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
			section, _ = c.file.GetSection("")
		} else {
			section, _ = c.file.GetSection(pathParts[0])
		}
		if section == nil {
			return nil, ErrorNotFound
		}
		key, _ := section.GetKey(pathParts[len(pathParts) - 1])
		if key == nil {
			return nil, ErrorNotFound
		}
		return key, nil
	}
	return nil, ErrorNotFound
}
