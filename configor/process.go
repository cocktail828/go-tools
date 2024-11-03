package configor

import (
	"fmt"
	"os"
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
		fieldStruct := configType.Field(i)
		field := configValue.Field(i)
		if !field.CanAddr() || !field.CanInterface() {
			continue
		}

		// read configuration from shell env
		if err := func() error {
			if !c.LoadEnv {
				return nil
			}

			envNames := []string{}
			if envName := fieldStruct.Tag.Get("env"); envName == "" {
				envName = strings.Join(append(prefixes, fieldStruct.Name), "_")
				envNames = append(envNames, envName, strings.ToUpper(envName))
			} else {
				envNames = []string{envName}
			}

			for _, name := range envNames {
				if c.EnvPrefix != "" {
					name = c.EnvPrefix + "_" + name
				}
				if value := os.Getenv(name); value != "" {
					switch reflect.Indirect(field).Kind() {
					case reflect.Bool:
						if val, err := strconv.ParseBool(strings.ToLower(value)); err == nil {
							field.Set(reflect.ValueOf(val))
						}
						return nil
					case reflect.String:
						field.Set(reflect.ValueOf(value))
						return nil
					default:
						return yaml.Unmarshal([]byte(value), field.Addr().Interface())
					}
				}
			}
			return nil
		}(); err != nil {
			return err
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
	data      []byte
	unmarshal Unmarshal
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
		if err := val.unmarshal(val.data, dst); err != nil {
			return err
		}
	}
	return c.processTags(dst)
}
