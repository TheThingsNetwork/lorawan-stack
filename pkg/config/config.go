// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package config wraps Viper. It also allows to set a struct with defaults and generates pflags
package config

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// TimeFormat is the format to parse times in.
const TimeFormat = time.RFC3339Nano

// Manager is a manager for the configuration.
type Manager struct {
	name       string
	viper      *viper.Viper
	flags      *pflag.FlagSet
	replacer   *strings.Replacer
	defaults   interface{}
	configFlag string
}

// Flags to be used in the command.
func (m *Manager) Flags() *pflag.FlagSet {
	return m.flags
}

// EnvKeyReplacer sets the strings.Replacer for mapping mapping an environmental variables to a key that does
// not match them.
func EnvKeyReplacer(r *strings.Replacer) Option {
	return func(m *Manager) {
		m.viper.SetEnvKeyReplacer(r)
		m.replacer = r
	}
}

// AllEnvironment returns all environment variables.
func (m *Manager) AllEnvironment() []string {
	keys := m.viper.AllKeys()
	env := make([]string, 0, len(keys))

	for _, key := range keys {
		env = append(env, strings.ToUpper(m.replacer.Replace(m.name+"."+key)))
	}

	return env
}

// AllKeys returns all keys holding a value, regardless of where they are set.
// Nested keys are returned with a "." separator.
func (m *Manager) AllKeys() []string {
	keys := m.viper.AllKeys()
	sort.Strings(keys)

	return keys
}

// Option is the type of an option for the manager.
type Option func(m *Manager)

// DefaultOptions are the default options.
var DefaultOptions = []Option{
	EnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_")),
	ConfigPath("config"),
}

// Initialize config with the given name and defaults.
func Initialize(name string, defaults interface{}, opts ...Option) *Manager {
	m := &Manager{
		name:     name,
		viper:    viper.New(),
		flags:    pflag.NewFlagSet(name, pflag.ExitOnError),
		replacer: strings.NewReplacer(),
		defaults: defaults,
	}

	m.viper.SetTypeByDefaultValue(true)
	m.viper.SetConfigName(name)
	m.viper.SetEnvPrefix(name)
	m.viper.AutomaticEnv()
	m.viper.AddConfigPath(".")

	m.flags.SetInterspersed(true)

	if defaults != nil {
		m.setDefaults("", defaults)
	}

	for _, opt := range append(DefaultOptions, opts...) {
		opt(m)
	}

	m.viper.BindPFlags(m.flags)

	return m
}

// Parse parses the command line arguments.
func (m *Manager) Parse(flags ...string) error {
	return m.flags.Parse(flags)
}

// Unmarshal unmarshals the available config keys into the result. It matches the names of fields based on the name struct tag.
func (m *Manager) Unmarshal(result interface{}) error {
	d, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          "name",
		ZeroFields:       true,
		WeaklyTypedInput: true,
		Result:           result,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			stringToTimeHookFunc(TimeFormat),
			stringSliceToStringMapHookFunc,
			stringSliceToStringHookFunc,
		),
	})

	if err != nil {
		return err
	}

	err = d.Decode(m.viper.AllSettings())
	if err != nil {
		return err
	}

	return nil
}

// ReadInConfig will load the configuration from disk. If a config file is set,
// that file will be used, otherwise ReadInConfig will discover the file.
func (m *Manager) ReadInConfig() error {
	files := m.viper.GetStringSlice(m.configFlag)
	for _, file := range files {
		m.viper.SetConfigFile(file)
		err := m.viper.MergeInConfig()
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) setDefaults(prefix string, config interface{}) {
	configValue := reflect.ValueOf(config)
	configKind := configValue.Type().Kind()

	if configKind == reflect.Interface || configKind == reflect.Ptr {
		configValue = configValue.Elem()
		configKind = configValue.Type().Kind()
	}

	if configKind != reflect.Struct {
		return
	}

	for i := 0; i < configValue.NumField(); i++ {
		field := configValue.Type().Field(i)
		name := field.Tag.Get("name")

		if name == "-" {
			continue
		}

		if name == "" {
			name = strings.ToLower(field.Name)
		}

		if prefix != "" {
			name = prefix + "." + name
		}

		description := field.Tag.Get("description")
		shorthand := field.Tag.Get("shorthand")

		if configValue.Field(i).CanInterface() {
			fieldType := field.Type.Kind()

			face := configValue.Field(i).Interface()
			if fieldType == reflect.Interface || fieldType == reflect.Ptr {
				if configValue.Field(i).IsNil() {
					continue
				}
				elem := configValue.Field(i).Elem()
				fieldType = elem.Type().Kind()
				face = elem.Interface()
			}

			switch val := face.(type) {
			case bool:
				m.viper.SetDefault(name, val)
				m.flags.BoolP(name, shorthand, val, description)

			case int, int8, int16, int32, int64:
				fieldValue := configValue.Field(i).Int()
				m.viper.SetDefault(name, int(fieldValue))
				m.flags.IntP(name, shorthand, int(fieldValue), description)

			case uint, uint8, uint16, uint32, uint64:
				fieldValue := configValue.Field(i).Uint()
				m.viper.SetDefault(name, uint(fieldValue))
				m.flags.UintP(name, shorthand, uint(fieldValue), description)

			case float32, float64:
				fieldValue := configValue.Field(i).Float()
				m.viper.SetDefault(name, float64(fieldValue))
				m.flags.Float64P(name, shorthand, float64(fieldValue), description)

			case string:
				m.viper.SetDefault(name, val)
				m.flags.StringP(name, shorthand, val, description)

			case time.Time:
				m.viper.SetDefault(name, val)
				m.flags.StringP(name, shorthand, val.Format(TimeFormat), description)

			case []string:
				m.viper.SetDefault(name, val)
				m.flags.StringSliceP(name, shorthand, val, description)

			case time.Duration:
				m.viper.SetDefault(name, val)
				m.flags.DurationP(name, shorthand, val, description)

			case map[string]string:
				defs := make([]string, 0, len(val))
				for k, v := range val {
					defs = append(defs, fmt.Sprintf("%s=%v", k, v))
				}

				m.flags.StringSliceP(name, shorthand, defs, description)
				m.viper.SetDefault(name, defs)

			default:
				switch fieldType {
				case reflect.Struct:
					if field.Anonymous {
						name = prefix
					}
					m.setDefaults(name, configValue.Field(i).Interface())
				default:
					panic(fmt.Errorf("config: cannot work with \"%v\" in configuration at name \"%s\"", field.Type, name))
				}
			}
		}
	}
}
