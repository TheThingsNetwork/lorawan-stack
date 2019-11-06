// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package config wraps Viper. It also allows to set a struct with defaults and generates pflags
package config

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// TimeFormat is the format to parse times in.
const TimeFormat = time.RFC3339Nano

// Manager is a manager for the configuration.
type Manager struct {
	name         string
	envPrefix    string
	viper        *viper.Viper
	flags        *pflag.FlagSet
	replacer     *strings.Replacer
	defaults     interface{}
	defaultPaths []string
	configFlag   string
	dataDirFlag  string
}

// Flags to be used in the command.
func (m *Manager) Flags() *pflag.FlagSet {
	return m.flags
}

// EnvKeyReplacer sets the strings.Replacer for mapping mapping an environment variables to a key that does
// not match them.
func EnvKeyReplacer(r *strings.Replacer) Option {
	return func(m *Manager) {
		m.viper.SetEnvKeyReplacer(r)
		m.replacer = r
	}
}

// AllEnvironment returns all environment variables.
func (m *Manager) AllEnvironment() []string {
	keys := m.AllKeys()
	env := make([]string, 0, len(keys))
	for _, key := range keys {
		env = append(env, m.EnvironmentForKey(key))
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

// EnvironmentForKey returns the name of the environment variable for the given config key.
func (m *Manager) EnvironmentForKey(key string) string {
	return strings.ToUpper(m.replacer.Replace(m.envPrefix + "." + key))
}

// Get returns the current value of the given config key.
func (m *Manager) Get(key string) interface{} {
	return m.viper.Get(key)
}

// Option is the type of an option for the manager.
type Option func(m *Manager)

func WithDeprecatedFlag(name, usageMessage string) Option {
	return func(m *Manager) {
		m.flags.MarkDeprecated(name, usageMessage)
	}
}

// DefaultOptions are the default options.
var DefaultOptions = []Option{
	EnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_")),
	WithConfigFileFlag("config"),
}

// Initialize a new config manager with the given name and defaults.
// defaults should be a struct wiath fields that define the possible config flags by setting the struct tags.
// Possible struct tags are:
//
//     `name:"<name>"`                Defines the name of the config flag, in the environment, on the command line and in the config files.
//     `shorthand:"<n>"`              Defines a shorthand name for use on the command line.
//     `description:"<description>"`  Add a description that will be printed in the command's help message.
//     `file-only:"<true|false>"`     Denotes wether or not to attempt to parse this variable from the command line and environment or only from the
//                                    config file. This can be used to allow complicated types to exist in the config file but not on the command line.
//
// The type of the struct fields also defines their type when parsing the config file, command line arguments or environment
// variables. Currently, the following types are supported:
//
//     bool
//     int, int8, int16, int32, int64
//     uint, uint8, uint16, uint32, uint64
//     float32, float64
//     string
//     time.Time                           Parsed according to the TimeFormat variable set in this package
//     time.Duration                       Parsed by time.ParseDuration
//     []string                            Parsed by splitting on whitespace or by passing multiple flags
//                                           VAR="a b c" or --var a --var b --var c
//     map[string]string                   Parsed by key=val pairs
//                                           VAR="k=v q=r" or --var k=v --var q=r
//     map[string][]byte                   Parsed by key=val pairs, val must be hex
//                                           VAR="k=0x01 q=0x02" or --var k=0x01 --var q=0x02
//     map[string][]string                 Parsed by key=val pairs where keys are repeated
//                                           VAR="k=v1 k=v2 q=r" or --var k=v1 --var k=v2 --var q=r
//     Configurable                        Parsed by the UnmarshalConfigString method
//     structs with fields of these types  The nested config names will be prefixed by the name of this struct, unless it is `name:",squash"`
//                                         in which case the names are merged into the parent struct.
func Initialize(name, envPrefix string, defaults interface{}, opts ...Option) *Manager {
	m := &Manager{
		name:      name,
		envPrefix: envPrefix,
		viper:     viper.New(),
		flags:     pflag.NewFlagSet(name, pflag.ExitOnError),
		replacer:  strings.NewReplacer(),
		defaults:  defaults,
	}

	m.viper.SetTypeByDefaultValue(true)
	m.viper.SetConfigName(name)
	m.viper.SetConfigType("yml")
	m.viper.AllowEmptyEnv(true)
	m.viper.SetEnvPrefix(envPrefix)
	m.viper.AutomaticEnv()
	m.viper.AddConfigPath(".")

	m.flags.SetInterspersed(true)

	if defaults != nil {
		m.setDefaults("", m.flags, defaults)
	}

	for _, opt := range opts {
		opt(m)
	}

	err := m.viper.BindPFlags(m.flags)
	if err != nil {
		panic(err)
	}

	return m
}

// WithConfig returns a new flagset with has the flags of the Manager as well as the additional flags defined
// from the defaults passed along.
// Use this to build derived flagsets with a shared base config (for instance with cobra).
func (m *Manager) WithConfig(defaults interface{}) *pflag.FlagSet {
	flags := pflag.NewFlagSet(m.name, pflag.ExitOnError)
	flags.AddFlagSet(m.flags)

	if defaults != nil {
		m.setDefaults("", flags, defaults)
	}

	err := m.viper.BindPFlags(flags)
	if err != nil {
		panic(err)
	}

	return flags
}

// InitializeWithDefaults is the same as Initialize but it sets some sane default options (see DefaultOptions)
// alongside the passed in options.
func InitializeWithDefaults(name, envPrefix string, defaults interface{}, opts ...Option) *Manager {
	return Initialize(name, envPrefix, defaults, append(DefaultOptions, opts...)...)
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
			stringSliceToStringMapStringSliceHookFunc,
			stringToStringMapHookFunc,
			stringToBufferMapHookFunc,
			stringSliceToStringHookFunc,
			configurableInterfaceHook,
			configurableInterfaceSliceHook,
			stringToByteSliceHook,
			stringToByteArrayHook,
		),
	})
	if err != nil {
		return err
	}

	return d.Decode(m.viper.AllSettings())
}

// the path must be in default paths
func (m *Manager) isDefault(path string) bool {
	for _, def := range m.defaultPaths {
		if def == path {
			return true
		}
	}

	return false
}

func (m *Manager) inCLIFlags(path string) bool {
	flags, err := m.flags.GetStringSlice(m.configFlag)
	if err != nil {
		return false
	}

	for _, flag := range flags {
		if path == flag {
			return true
		}
	}
	return false
}

// ReadInConfig will read in all defined config files (according to the config file flag set by WithConfigFileFlag).
// The parsed config files will be merged into the config struct.
func (m *Manager) ReadInConfig() error {
	files := m.viper.GetStringSlice(m.configFlag)
	for _, file := range files {
		// ignore default config files that do not exist
		if m.isDefault(file) && !m.inCLIFlags(file) {
			if _, err := os.Stat(file); os.IsNotExist(err) {
				continue
			}
		}

		m.viper.SetConfigFile(file)
		err := m.viper.MergeInConfig()
		if err != nil {
			return err
		}
	}

	return nil
}

// mergeConfig merges the config from the reader as a yml config file.
func (m *Manager) mergeConfig(in io.Reader) error {
	return m.viper.MergeConfig(in)
}

// UnmarshalKey unmarshals a specific key into a destination, which must have a matching type.
// This is useful for fields which have the `file-only:"true"` tag set and so are ignored when
// Unmarshalling them to a struct.
func (m *Manager) UnmarshalKey(key string, raw interface{}) error {
	return m.viper.UnmarshalKey(key, raw)
}

// Configurable is the interface for things that can be configured.
// Implement the interface to add custom parsing to config variables from strings.
// For instance, to parse a log level from the strings "fatal", "error", etc into a custom
// enum for log levels.
type Configurable interface {
	// UnmarshalConfigString parses a string into the config variable
	UnmarshalConfigString(string) error
}

// Stringer is the interface for config variables that have a custom string representation.
// Implement next to Configurable if you want custom parsing and formatting for a type, and if the formatting
// needs to be different from fmt.String for some reason.
type Stringer interface {
	// ConfigString returns the config string representation of type
	ConfigString() string
}

var configurableI = reflect.TypeOf((*Configurable)(nil)).Elem()

func isConfigurableType(t reflect.Type) bool {
	return t.Implements(configurableI) || reflect.PtrTo(t).Implements(configurableI)
}

func (m *Manager) setDefaults(prefix string, flags *pflag.FlagSet, config interface{}) {
	configValue := reflect.ValueOf(config)
	configKind := configValue.Type().Kind()

	if configKind == reflect.Interface || configKind == reflect.Ptr {
		configValue = configValue.Elem()
		configKind = configValue.Type().Kind()
	}

	if configKind != reflect.Struct {
		panic("default config is not a struct type")
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

		// skip previously defined flags
		if f := flags.Lookup(name); f != nil {
			continue
		}

		description := field.Tag.Get("description")
		shorthand := field.Tag.Get("shorthand")
		fileOnly := field.Tag.Get("file-only")

		if configValue.Field(i).CanInterface() {
			fieldKind := field.Type.Kind()

			face := configValue.Field(i).Interface()

			// if it's only for in the file, skip the rest
			if fileOnly == "true" {
				m.viper.SetDefault(name, face)
				continue
			}

			if isConfigurableType(field.Type) {
				val := fmt.Sprintf("%v", face)

				if str, ok := face.(fmt.Stringer); ok {
					val = str.String()
				}

				if cstr, ok := face.(Stringer); ok {
					val = cstr.ConfigString()
				}

				m.viper.SetDefault(name, val)
				m.flags.StringP(name, shorthand, val, description)
				continue
			}

			if fieldKind == reflect.Slice && isConfigurableType(field.Type.Elem()) {
				val := configValue.Field(i)
				n := val.Len()
				defs := make([]string, 0, n)

				for j := 0; j < n; j++ {
					c := val.Index(j).Interface()
					str := fmt.Sprintf("%v", c)

					if s, ok := c.(fmt.Stringer); ok {
						str = s.String()
					}

					if s, ok := c.(Stringer); ok {
						str = s.ConfigString()
					}

					defs = append(defs, str)
				}

				m.viper.SetDefault(name, defs)
				m.flags.StringSliceP(name, shorthand, defs, description)
				continue
			}

			if fieldKind == reflect.Interface || fieldKind == reflect.Ptr {
				if configValue.Field(i).IsNil() {
					continue
				}
				elem := configValue.Field(i).Elem()
				fieldKind = elem.Type().Kind()
				face = elem.Interface()
			}

			switch val := face.(type) {
			case bool:
				m.viper.SetDefault(name, val)
				flags.BoolP(name, shorthand, val, description)

			case int, int8, int16, int32, int64:
				fieldValue := reflect.Indirect(configValue.Field(i)).Int()
				m.viper.SetDefault(name, int(fieldValue))
				flags.IntP(name, shorthand, int(fieldValue), description)

			case uint, uint8, uint16, uint32, uint64:
				fieldValue := reflect.Indirect(configValue.Field(i)).Uint()
				m.viper.SetDefault(name, uint(fieldValue))
				flags.UintP(name, shorthand, uint(fieldValue), description)

			case float32, float64:
				fieldValue := reflect.Indirect(configValue.Field(i)).Float()
				m.viper.SetDefault(name, fieldValue)
				flags.Float64P(name, shorthand, fieldValue, description)

			case string:
				m.viper.SetDefault(name, val)
				flags.StringP(name, shorthand, val, description)

			case time.Time:
				m.viper.SetDefault(name, val)
				flags.StringP(name, shorthand, val.Format(TimeFormat), description)

			case []string:
				m.viper.SetDefault(name, val)
				flags.StringSliceP(name, shorthand, val, description)

			case time.Duration:
				m.viper.SetDefault(name, val)
				flags.DurationP(name, shorthand, val, description)

			case map[string]string:
				defs := make([]string, 0, len(val))
				for k, v := range val {
					defs = append(defs, fmt.Sprintf("%s=%v", k, v))
				}

				flags.StringSliceP(name, shorthand, defs, description)
				m.viper.SetDefault(name, val)

			case map[string][]byte:
				defs := make([]string, 0, len(val))
				for k, v := range val {
					if len(v) > 0 {
						defs = append(defs, fmt.Sprintf("%s=0x%X", k, v))
					}
				}

				flags.StringSliceP(name, shorthand, defs, description)
				m.viper.SetDefault(name, val)

			case map[string][]string:
				defs := make([]string, 0, len(val))
				for k, vs := range val {
					for _, v := range vs {
						defs = append(defs, fmt.Sprintf("%s=%v", k, v))
					}
				}

				flags.StringSliceP(name, shorthand, defs, description)
				m.viper.SetDefault(name, val)

			case []byte:
				var str string
				if len(val) > 0 {
					str = fmt.Sprintf("0x%X", val)
					m.viper.SetDefault(name, str)
				}
				flags.StringP(name, shorthand, str, description)

			case types.NetID:
				str := val.String()
				m.viper.SetDefault(name, str)
				flags.StringP(name, shorthand, str, description)

			case ttnpb.RxDelay:
				m.viper.SetDefault(name, int32(val))
				flags.Int32P(name, shorthand, int32(val), description)

			default:
				switch fieldKind {
				case reflect.Struct:
					if field.Anonymous {
						name = prefix
					}
					m.setDefaults(name, flags, configValue.Field(i).Interface())
				default:
					panic(fmt.Errorf("config: cannot work with \"%v\" in configuration at name \"%s\"", field.Type, name))
				}
			}
		}
	}
}
