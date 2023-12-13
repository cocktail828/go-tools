package configor

import (
	"os"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

func Load(dst interface{}, files ...string) error {
	return newConfigor().Load(dst, files...)
}

func LoadContents(dst interface{}, data ...string) error {
	return newConfigor().LoadContent(dst, data...)
}

type configor struct {
	ENVPrefix string
	// In case of json files, this field will be used only when compiled with
	// go 1.10 or later.
	// This field will be ignored when compiled with go versions lower than 1.10.
	ErrorOnUnmatchedKeys bool
}

func newConfigor() *configor {
	return &configor{}
}

// GetErrorOnUnmatchedKeys returns a boolean indicating if an error should be
// thrown if there are keys in the dst file that do not correspond to the
// dst struct
func (c *configor) GetErrorOnUnmatchedKeys() bool {
	return c.ErrorOnUnmatchedKeys
}

// Load will unmarshal configurations to struct from files that you provide
func (c *configor) Load(dst interface{}, files ...string) error {
	for _, file := range files {
		if fileInfo, err := os.Stat(file); err != nil || fileInfo.Mode().IsRegular() {
			return errors.Errorf("Failed to find configuration %v", file)
		}
		if err := processFile(dst, file, c.GetErrorOnUnmatchedKeys()); err != nil {
			return err
		}
	}
	return c.processTags(dst)
}

// LoadContent will unmarshal configurations to struct from data that you provide
func (c *configor) LoadContent(dst interface{}, data ...string) error {
	for _, content := range data {
		if err := unmarshalTomlString(content, dst, c.GetErrorOnUnmatchedKeys()); err != nil {
			return err
		}
	}
	return c.processTags(dst)
}

func (c *configor) processTags(config interface{}) error {
	configValue := reflect.Indirect(reflect.ValueOf(config))
	if configValue.Kind() != reflect.Struct {
		return errors.New("invalid config, should be struct")
	}

	configType := configValue.Type()
	for i := 0; i < configType.NumField(); i++ {
		var (
			envNames    []string
			fieldStruct = configType.Field(i)
			field       = configValue.Field(i)
			envName     = fieldStruct.Tag.Get("env") // read configuration from shell env
		)

		if !field.CanAddr() || !field.CanInterface() {
			continue
		}

		if envName == "" {
			envNames = append(envNames, fieldStruct.Name, strings.ToUpper(fieldStruct.Name))
		} else {
			envNames = []string{envName}
		}

		// Load From Shell ENV
		for _, env := range envNames {
			if value := os.Getenv(env); value != "" {
				if err := yaml.Unmarshal([]byte(value), field.Addr().Interface()); err != nil {
					return err
				}
				break
			}
		}

		if isBlank := reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()); isBlank {
			// Set default configuration if blank
			if value := fieldStruct.Tag.Get("default"); value != "" {
				if err := yaml.Unmarshal([]byte(value), field.Addr().Interface()); err != nil {
					return err
				}
			} else if fieldStruct.Tag.Get("required") == "true" {
				// return error if it is required but blank
				return errors.New(fieldStruct.Name + " is required, but blank")
			}
		}

		for field.Kind() == reflect.Ptr {
			field = field.Elem()
		}

		if field.Kind() == reflect.Struct {
			if err := c.processTags(field.Addr().Interface()); err != nil {
				return err
			}
		}

		if field.Kind() == reflect.Slice {
			for i := 0; i < field.Len(); i++ {
				if reflect.Indirect(field.Index(i)).Kind() == reflect.Struct {
					if err := c.processTags(field.Index(i).Addr().Interface()); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}
