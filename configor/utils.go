package configor

import (
	"fmt"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"
)

func (c *Configor) getPrefixForStruct(prefixes []string, fieldStruct *reflect.StructField) []string {
	if fieldStruct.Anonymous && fieldStruct.Tag.Get("anonymous") == "true" {
		return prefixes
	}
	return append(prefixes, fieldStruct.Name)
}

func (c *Configor) processDefaults(dst any) error {
	configValue := reflect.Indirect(reflect.ValueOf(dst))
	if configValue.Kind() != reflect.Struct {
		return errors.New("invalid dst, should be struct")
	}

	configType := configValue.Type()
	for i := 0; i < configType.NumField(); i++ {
		var (
			fieldStruct = configType.Field(i)
			field       = configValue.Field(i)
		)

		if !field.CanAddr() || !field.CanInterface() {
			continue
		}

		if isBlank := reflect.DeepEqual(field.Interface(), reflect.Zero(field.Type()).Interface()); isBlank {
			// Set default configuration if blank
			if value := fieldStruct.Tag.Get("default"); value != "" {
				if err := yaml.Unmarshal([]byte(value), field.Addr().Interface()); err != nil {
					return err
				}
			}
		}

		for field.Kind() == reflect.Ptr {
			field = field.Elem()
		}

		switch field.Kind() {
		case reflect.Struct:
			if err := c.processDefaults(field.Addr().Interface()); err != nil {
				return err
			}
		case reflect.Slice:
			for i := 0; i < field.Len(); i++ {
				if reflect.Indirect(field.Index(i)).Kind() == reflect.Struct {
					if err := c.processDefaults(field.Index(i).Addr().Interface()); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (c *Configor) processTags(config interface{}, prefixes ...string) error {
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
			envNames = append(envNames, strings.Join(append(prefixes, fieldStruct.Name), "_"))                  // Configor_DB_Name
			envNames = append(envNames, strings.ToUpper(strings.Join(append(prefixes, fieldStruct.Name), "_"))) // CONFIGOR_DB_NAME
		} else {
			envNames = []string{envName}
		}

		// Load From Shell ENV
	loop:
		for _, env := range envNames {
			name := env
			if c.EnvPrefix != "" {
				name = c.EnvPrefix + "_" + env
			}
			if value := os.Getenv(name); value != "" {
				switch reflect.Indirect(field).Kind() {
				case reflect.Bool:
					if val, err := strconv.ParseBool(strings.ToLower(value)); err == nil {
						field.Set(reflect.ValueOf(val))
					}
					break loop
				case reflect.String:
					field.Set(reflect.ValueOf(value))
					break loop
				default:
					if err := yaml.Unmarshal([]byte(value), field.Addr().Interface()); err != nil {
						return err
					} else {
						break loop
					}
				}
			}
		}

		for field.Kind() == reflect.Ptr {
			field = field.Elem()
		}

		if field.Kind() == reflect.Struct {
			if err := c.processTags(field.Addr().Interface(), c.getPrefixForStruct(prefixes, &fieldStruct)...); err != nil {
				return err
			}
		}

		if field.Kind() == reflect.Slice {
			if arrLen := field.Len(); arrLen > 0 {
				for i := 0; i < arrLen; i++ {
					if reflect.Indirect(field.Index(i)).Kind() == reflect.Struct {
						if err := c.processTags(field.Index(i).Addr().Interface(), append(c.getPrefixForStruct(prefixes, &fieldStruct), fmt.Sprint(i))...); err != nil {
							return err
						}
					}
				}
			} else {
				defer func(field reflect.Value, fieldStruct reflect.StructField) {
					if !configValue.IsZero() {
						// load slice from env
						newVal := reflect.New(field.Type().Elem()).Elem()
						if newVal.Kind() == reflect.Struct {
							idx := 0
							for {
								newVal = reflect.New(field.Type().Elem()).Elem()
								if err := c.processTags(newVal.Addr().Interface(), append(c.getPrefixForStruct(prefixes, &fieldStruct), fmt.Sprint(idx))...); err != nil {
									return // err
								} else if reflect.DeepEqual(newVal.Interface(), reflect.New(field.Type().Elem()).Elem().Interface()) {
									break
								} else {
									idx++
									field.Set(reflect.Append(field, newVal))
								}
							}
						}
					}
				}(field, fieldStruct)
			}
		}
	}
	return nil
}

type pair struct {
	payload     []byte
	unmarshaler func([]byte, any) error
}

func (c *Configor) loadFile(dst any, files ...string) error {
	pairs := make([]pair, 0, len(files))
	for _, fname := range files {
		data, err := os.ReadFile(fname)
		if err != nil {
			return err
		}
		if f, ok := unmarshalers[path.Ext(fname)]; ok {
			pairs = append(pairs, pair{data, f})
		} else {
			pairs = append(pairs, pair{data, c.Unmarshaler})
		}
	}
	return c.internalLoad(dst, pairs...)
}

func (c *Configor) load(dst any, payloads ...[]byte) error {
	pairs := make([]pair, 0, len(payloads))
	for _, body := range payloads {
		pairs = append(pairs, pair{body, c.Unmarshaler})
	}
	return c.internalLoad(dst, pairs...)
}

func (c *Configor) internalLoad(dst any, pairs ...pair) error {
	defaultValue := reflect.Indirect(reflect.ValueOf(dst))
	if !defaultValue.CanAddr() {
		return errors.Errorf("Config %v should be addressable", dst)
	}
	if err := c.processDefaults(dst); err != nil {
		return err
	}
	for _, val := range pairs {
		if err := val.unmarshaler(val.payload, dst); err != nil {
			return err
		}
	}
	return c.processTags(dst)
}
