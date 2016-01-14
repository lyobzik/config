package config

import (
	ini "gopkg.in/ini.v1"
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
	return GrabStringValue(c, path, createIniValueGrabber(grabber))
}

func (c *iniConfig) GrabValues(path string, delim string,
	creator ValueSliceCreator, grabber ValueGrabber) (err error) {

	return GrabStringValues(c, path, delim, creator, createIniValueGrabber(grabber))
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
			if parsed, err = parseIniBool(data); err == nil {
				value = append(value, parsed)
			}
			return err
		})
}

func (c *iniConfig) GetFloats(path string, delim string) (value []float64, err error) {
	return value, GrabStringValues(c, path, delim,
		func(cap int) { value = make([]float64, 0, cap) },
		func(data string) error {
			var parsed float64
			if parsed, err = parseIniFloat(data); err == nil {
				value = append(value, parsed)
			}
			return err
		})
}

func (c *iniConfig) GetInts(path string, delim string) (value []int64, err error) {
	return value, GrabStringValues(c, path, delim,
		func(cap int) { value = make([]int64, 0, cap) },
		func(data string) error {
			var parsed int64
			if parsed, err = parseIniInt(data); err == nil {
				value = append(value, parsed)
			}
			return err
		})
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
			return nil, ErrorNotFound
		}
		return c.key, nil
	}

	if c.section != nil {
		if len(pathParts) != 1 {
			return nil, ErrorNotFound
		}
		key, _ := c.section.GetKey(pathParts[0])
		if key == nil {
			return nil, ErrorNotFound
		}
		return key, nil
	}

	if c.file != nil {
		if len(pathParts) < 1 || 2 < len(pathParts) {
			return nil, ErrorNotFound
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

// Ini value parsers.
func parseIniBool(data string) (bool, error) {
	return parseXmlBool(data)
}

func parseIniFloat(data string) (float64, error) {
	return parseXmlFloat(data)
}

func parseIniInt(data string) (int64, error) {
	return parseXmlInt(data)
}

// Grabbing helpers.
func createIniValueGrabber(grabber ValueGrabber) StringValueGrabber {
	return createXmlValueGrabber(grabber)
}
