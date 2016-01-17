package config

import (
	ini "gopkg.in/ini.v1"
)

type iniConfig struct {
	file    *ini.File
	section *ini.Section
	key     *ini.Key
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
	section, key, err := c.findElement(path)
	if err != nil {
		return nil, err
	}
	if section == nil && key == nil {
		return &iniConfig{file: c.file}, nil
	}
	return &iniConfig{section: section, key: key}, nil
}

// Ini helpers.
func (c *iniConfig) findKey(path string) (*ini.Key, error) {
	_, key, err := c.findElement(path)
	if err != nil {
		return nil, err
	}
	if key == nil {
		return nil, ErrorNotFound
	}
	return key, nil
}

func (c *iniConfig) findElement(path string) (*ini.Section, *ini.Key, error) {
	pathParts := splitPath(path)
	section, pathParts, err := c.getSection(c.file, pathParts)
	if err != nil {
		return nil, nil, err
	}
	key, pathParts, err := c.getKey(section, pathParts)
	if err != nil {
		return nil, nil, err
	}
	if len(pathParts) > 0 {
		return nil, nil, ErrorNotFound
	}
	return section, key, nil
}

func (c *iniConfig) getSection(file *ini.File, path []string) (*ini.Section, []string, error) {
	if c.section != nil {
		return c.section, path, nil
	}
	if file != nil && len(path) > 0 {
		var section *ini.Section
		if len(file.Sections()) == 1 {
			section, _ = file.GetSection("")
		} else if len(path) > 0 {
			section, _ = file.GetSection(path[0])
			path = path[1:]
		}
		if section == nil {
			return nil, nil, ErrorNotFound
		}
		return section, path, nil
	}
	return nil, path, nil
}

func (c *iniConfig) getKey(section *ini.Section, path []string) (*ini.Key, []string, error) {
	if c.key != nil {
		return c.key, path, nil
	}
	if section != nil && len(path) > 0 {
		key, _ := section.GetKey(path[0])
		if key == nil {
			return nil, nil, ErrorNotFound
		}
		return key, path[1:], nil
	}
	return nil, path, nil
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
