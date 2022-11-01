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

package util

import (
	"encoding"
	"encoding/hex"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// NormalizedFlagSet returns a flagset with a NormalizeFunc that replaces underscores to dashes.
func NormalizedFlagSet() *pflag.FlagSet {
	fs := &pflag.FlagSet{}
	fs.SetNormalizeFunc(NormalizeFlags)
	return fs
}

// DeprecateFlag deprecates a CLI flag.
func DeprecateFlag(flagSet *pflag.FlagSet, old string, new string) {
	if newFlag := flagSet.Lookup(new); newFlag != nil {
		deprecated := *newFlag
		deprecated.Name = old
		deprecated.Usage = strings.Replace(deprecated.Usage, old, new, -1)
		deprecated.Deprecated = fmt.Sprintf("use the %s flag", new)
		deprecated.Hidden = true
		flagSet.AddFlag(&deprecated)
	}
}

// ForwardFlag forwards the flag value of old to new if new is not set while old is.
func ForwardFlag(flagSet *pflag.FlagSet, old string, new string) {
	if oldFlag := flagSet.Lookup(old); oldFlag != nil && oldFlag.Changed {
		if newFlag := flagSet.Lookup(new); newFlag != nil && !newFlag.Changed {
			flagSet.Set(new, oldFlag.Value.String())
		}
	}
}

// HideFlag hides the provided flag from the flag set.
func HideFlag(flagSet *pflag.FlagSet, name string) {
	if flag := flagSet.Lookup(name); flag != nil {
		flag.Hidden = true
	}
}

// HideFlagSet hides the flags from the provided flag set.
func HideFlagSet(flagSet *pflag.FlagSet) *pflag.FlagSet {
	flagSet.VisitAll(func(f *pflag.Flag) {
		f.Hidden = true
	})
	return flagSet
}

var (
	toDash       = strings.NewReplacer("_", "-")
	toUnderscore = strings.NewReplacer("-", "_")
)

// NormalizePaths converts arguments to field mask paths, replacing '-' with '_'
func NormalizePaths(paths []string) []string {
	normalized := make([]string, len(paths))
	for i, path := range paths {
		normalized[i] = toUnderscore.Replace(path)
	}
	return normalized
}

func NormalizeFlags(f *pflag.FlagSet, name string) pflag.NormalizedName {
	return pflag.NormalizedName(toDash.Replace(name))
}

func SelectFieldMask(cmdFlags *pflag.FlagSet, fieldMaskFlags ...*pflag.FlagSet) (paths []string) {
	if all, _ := cmdFlags.GetBool("all"); all {
		for _, fieldMaskFlags := range fieldMaskFlags {
			fieldMaskFlags.VisitAll(func(flag *pflag.Flag) {
				paths = append(paths, toUnderscore.Replace(flag.Name))
			})
		}
		return
	}
	cmdFlags.Visit(func(flag *pflag.Flag) {
		flagName := toUnderscore.Replace(flag.Name)
		for _, fieldMaskFlags := range fieldMaskFlags {
			if b, err := fieldMaskFlags.GetBool(flag.Name); err == nil && b {
				paths = append(paths, flagName)
				return
			}
		}
	})
	return
}

func UpdateFieldMask(cmdFlags *pflag.FlagSet, fieldMaskFlags ...*pflag.FlagSet) (paths []string) {
	cmdFlags.Visit(func(flag *pflag.Flag) {
		flagName := toUnderscore.Replace(flag.Name)
		for _, fieldMaskFlags := range fieldMaskFlags {
			if fieldMaskFlags.Lookup(flagName) != nil {
				paths = append(paths, flagName)
				return
			}
		}
	})
	return
}

func FieldFlags(v interface{}, prefix ...string) *pflag.FlagSet {
	t := reflect.Indirect(reflect.ValueOf(v)).Type()
	return fieldMaskFlags(prefix, t, false)
}

func FieldMaskFlags(v interface{}, prefix ...string) *pflag.FlagSet {
	t := reflect.Indirect(reflect.ValueOf(v)).Type()
	return fieldMaskFlags(prefix, t, true)
}

func isStopType(t reflect.Type, maskOnly bool) bool {
	switch t.PkgPath() {
	case "github.com/gogo/protobuf/types":
		switch t.Name() {
		case
			// Struct is a standalone type and should not be parsed further for field mask purposes.
			"Struct":
			return true
		}
	case "go.thethings.network/lorawan-stack/v3/pkg/ttnpb":
		switch t.Name() {
		case
			// ErrorDetails is a recursive type, stop parsing to prevent infinite recursion.
			"ErrorDetails":
			return true
		default:
		}
	}
	return false
}

func isAtomicType(t reflect.Type, maskOnly bool) bool {
	switch t.PkgPath() {
	case "time":
		switch t.Name() {
		case "Time", "Duration":
			return true
		}
	case "github.com/gogo/protobuf/types":
		switch t.Name() {
		case
			"BoolValue",
			"BytesValue",
			"DoubleValue",
			"FloatValue",
			"Int32Value",
			"Int64Value",
			"StringValue",
			"UInt32Value",
			"UInt64Value",
			"Timestamp",
			"Duration":
			return true
		}
	case "go.thethings.network/lorawan-stack/v3/pkg/ttnpb":
		switch t.Name() {
		case
			"ADRAckDelayExponentValue",
			"ADRAckLimitExponentValue",
			"AggregatedDutyCycleValue",
			"BoolValue",
			"DataRateIndexValue",
			"DataRateOffsetValue",
			"DeviceEIRPValue",
			"FrequencyValue",
			"ZeroableFrequencyValue",
			"GatewayAntennaIdentifiers",
			"Picture",
			"PingSlotPeriodValue",
			"RxDelayValue":
			return true
		}
	}
	return false
}

func isSelectableField(name string) bool {
	switch name {
	case "created_at", "updated_at", "ids":
		return false
	}
	return true
}

func isSettableField(name string) bool {
	switch name {
	case "attributes", "contact_info", "password_updated_at", "temporary_password_created_at", "antennas", "profile_picture":
		return false
	}
	return true
}

func enumValues(t reflect.Type) []string {
	if t.PkgPath() == "go.thethings.network/lorawan-stack/v3/pkg/ttnpb" {
		valueMap := make(map[string]int32)
		implementsStringer := t.Implements(reflect.TypeOf((*fmt.Stringer)(nil)).Elem())
		for s, v := range proto.EnumValueMap(fmt.Sprintf("ttn.lorawan.v3.%s", t.Name())) {
			valueMap[s] = v
			if implementsStringer {
				// If the enum implements Stringer, then the String might be different than the official name.
				rv := reflect.New(t).Elem()
				rv.SetInt(int64(v))
				valueMap[rv.Interface().(fmt.Stringer).String()] = v
			}
		}
		values := make([]string, 0, len(valueMap))
		for value := range valueMap {
			values = append(values, value)
		}
		sort.Strings(values)
		return values
	}
	return nil
}

func unwrapLoRaWANEnumType(typeName string) string {
	return fmt.Sprintf("ttn.lorawan.v3.%s", strings.TrimSuffix(strings.TrimPrefix(typeName, ""), "Value"))
}

// AddField adds a field to the flag set.
func AddField(fs *pflag.FlagSet, name string, t reflect.Type, maskOnly bool) {
	if maskOnly {
		if t.Kind() == reflect.Struct && !isAtomicType(t, maskOnly) {
			fs.Bool(name, false, fmt.Sprintf("select the %s field and all allowed sub-fields", name))
			return
		}
		fs.Bool(name, false, fmt.Sprintf("select the %s field", name))
		return
	}

	if t.Kind() == reflect.Struct && !isAtomicType(t, maskOnly) {
		return
	}

	switch t.PkgPath() {
	case "time":
		switch t.Name() {
		case "Time":
			fs.String(name, "", "(YYYY-MM-DDTHH:MM:SSZ)")
			return
		case "Duration":
			fs.Duration(name, 0, "(1h2m3s)")
			return
		}

	case "github.com/gogo/protobuf/types":
		switch t.Name() {
		case "DoubleValue":
			fs.Float64(name, 0, "")
			return
		case "FloatValue":
			fs.Float32(name, 0, "")
			return
		case "Int64Value":
			fs.Int64(name, 0, "")
			return
		case "UInt64Value":
			fs.Uint64(name, 0, "")
			return
		case "Int32Value":
			fs.Int32(name, 0, "")
			return
		case "UInt32Value":
			fs.Uint32(name, 0, "")
			return
		case "BoolValue":
			fs.Bool(name, false, "")
			return
		case "StringValue":
			fs.String(name, "", "")
			return
		case "BytesValue":
			fs.String(name, "", "(hex)")
			return
		case "Timestamp":
			fs.String(name, "", "(YYYY-MM-DDTHH:MM:SSZ)")
			return
		case "Duration":
			fs.Duration(name, 0, "(1h2m3s)")
			return
		}

	case "go.thethings.network/lorawan-stack/v3/pkg/ttnpb":
		switch typeName := t.Name(); typeName {
		case
			"ADRAckDelayExponentValue",
			"ADRAckLimitExponentValue",
			"AggregatedDutyCycleValue",
			"DataRateIndexValue",
			"DataRateOffsetValue",
			"DeviceEIRPValue",
			"PingSlotPeriodValue",
			"RxDelayValue":
			fv, ok := t.FieldByName("Value")
			if !ok {
				panic(fmt.Sprintf("flags: %T type does not contain a Value field", typeName))
			}
			values := enumValues(fv.Type)
			if len(values) == 0 {
				panic(fmt.Sprintf("flags: no allowed values for %T", typeName))
			}
			fs.String(name, "", fmt.Sprintf("allowed values: %s", strings.Join(values, ", ")))
			return

		case "Picture":
			// Not supported
			return

		case "FrequencyValue", "ZeroableFrequencyValue":
			fs.Uint64(name, 0, "")
			return

		case "BoolValue":
			fs.Bool(name, false, "")
			return
		}
		if t.Kind() == reflect.Int32 {
			if values := enumValues(t); values != nil {
				fs.String(name, "", fmt.Sprintf("allowed values: %s", strings.Join(values, ", ")))
				return
			}
		}
	}

	switch t.Kind() {
	case reflect.Bool:
		fs.Bool(name, false, "")
		return
	case reflect.String:
		fs.String(name, "", "")
		return
	case reflect.Int32:
		fs.Int32(name, 0, "")
		return
	case reflect.Int64:
		fs.Int64(name, 0, "")
		return
	case reflect.Uint32:
		fs.Uint32(name, 0, "")
		return
	case reflect.Uint64:
		fs.Uint64(name, 0, "")
		return
	case reflect.Float32:
		fs.Float32(name, 0, "")
		return
	case reflect.Float64:
		fs.Float64(name, 0, "")
		return
	case reflect.Slice:
		el := t.Elem()

		if el.Kind() == reflect.Ptr {
			el := el.Elem()

			switch el.PkgPath() {
			case "go.thethings.network/lorawan-stack/v3/pkg/ttnpb":
				switch el.Name() {
				case "GatewayAntennaIdentifiers":
					fs.StringSlice(name, nil, "")
					return
				}
			}
		}

		switch el.PkgPath() {
		case "go.thethings.network/lorawan-stack/v3/pkg/types":
			switch el.Name() {
			case "DevAddrPrefix":
				fs.StringSlice(name, nil, "")
				return
			}
		}

		switch el.Kind() {
		case reflect.Bool:
			fs.BoolSlice(name, nil, "")
			return
		case reflect.String:
			fs.StringSlice(name, nil, "")
			return
		case reflect.Int32:
			if values := enumValues(el); values != nil {
				fs.StringSlice(name, nil, fmt.Sprintf("allowed values: %s", strings.Join(values, ", ")))
				return
			}
			fs.IntSlice(name, nil, "")
			return
		case reflect.Int64:
			fs.IntSlice(name, nil, "")
			return
		case reflect.Uint8:
			fs.String(name, "", "(hex)")
			return
		case reflect.Uint32, reflect.Uint64:
			fs.UintSlice(name, nil, "")
			return
		case reflect.Ptr:
			// Not supported
			return
		}
		panic(fmt.Sprintf("flags: %s slice not yet supported (%s)\n", el.Kind(), name))
	case reflect.Array:
		el := t.Elem()
		switch el.Kind() {
		case reflect.Uint8:
			fs.String(name, "", "(hex)")
			return
		}
		panic(fmt.Sprintf("flags: %s array not yet supported (%s)\n", el.Kind(), name))
	case reflect.Map:
		// Not supported
		return
	}
	if t.PkgPath() == "" {
		panic(fmt.Sprintf("flags: %s not yet supported (%s)\n", t.Kind(), name))
	}
	panic(fmt.Sprintf("flags: %s.%s not yet supported (%s)\n", t.PkgPath(), t.Name(), name))
}

func fieldMaskFlags(prefix []string, t reflect.Type, maskOnly bool) *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	props := proto.GetProperties(t)
	for _, oneofProp := range props.OneofTypes {
		if oneofProp.Prop.Tag == 0 {
			continue
		}

		fieldType := oneofProp.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}

		path := append(prefix, props.Prop[oneofProp.Field].OrigName)

		if fieldType.Kind() == reflect.Struct && !isAtomicType(fieldType, maskOnly) {
			flagSet.AddFlagSet(fieldMaskFlags(path, fieldType, maskOnly))
		}
	}
	for _, prop := range props.Prop {
		if prop.Tag == 0 {
			continue
		}
		field, ok := t.FieldByName(prop.Name)
		if !ok {
			continue
		}
		path := append(prefix, prop.OrigName)
		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			fieldType = fieldType.Elem()
		}
		name := strings.Join(path, ".")
		if !isSelectableField(name) {
			continue
		}
		if !maskOnly && !isSettableField(name) {
			continue
		}
		if isStopType(t, maskOnly) {
			continue
		}
		AddField(flagSet, name, fieldType, maskOnly)
		if fieldType.Kind() == reflect.Struct && !isAtomicType(fieldType, maskOnly) {
			flagSet.AddFlagSet(fieldMaskFlags(path, fieldType, maskOnly))
		}
	}
	return flagSet
}

func trimPrefix(path []string, prefix ...string) []string {
	var nextElement int
	for i, prefix := range prefix {
		if i >= len(path) || path[i] != prefix {
			break
		}
		nextElement = i + 1
	}
	return path[nextElement:]
}

var flagValueError = errors.DefineInvalidArgument("flag_value", "invalid flag value")

func SetFields(dst interface{}, flags *pflag.FlagSet, prefix ...string) error {
	var flagValueErrorAttributes []interface{}
	rv := reflect.Indirect(reflect.ValueOf(dst))
	flags.VisitAll(func(flag *pflag.Flag) {
		if flag.Deprecated != "" || !flag.Changed {
			return
		}
		flagName := toUnderscore.Replace(flag.Name)
		var v interface{}
		switch flag.Value.Type() {
		case "bool":
			v, _ = flags.GetBool(flagName)
		case "string":
			v, _ = flags.GetString(flagName)
		case "int32":
			v, _ = flags.GetInt32(flagName)
		case "int64":
			v, _ = flags.GetInt64(flagName)
		case "uint32":
			v, _ = flags.GetUint32(flagName)
		case "uint64":
			v, _ = flags.GetUint64(flagName)
		case "float32":
			v, _ = flags.GetFloat32(flagName)
		case "float64":
			v, _ = flags.GetFloat64(flagName)
		case "stringSlice":
			v, _ = flags.GetStringSlice(flagName)
		case "intSlice":
			v, _ = flags.GetIntSlice(flagName)
		case "uintSlice":
			v, _ = flags.GetUintSlice(flagName)
		case "duration":
			v, _ = flags.GetDuration(flagName)
		}
		if v == nil {
			flagValueErrorAttributes = append(flagValueErrorAttributes,
				flag.Name, fmt.Errorf("can't set field to %s (%v)", flag.Value, flag.Value.Type()),
			)
		}
		if err := setField(rv, trimPrefix(strings.Split(flagName, "."), prefix...), reflect.ValueOf(v)); err != nil {
			flagValueErrorAttributes = append(flagValueErrorAttributes,
				flag.Name, err.Error(),
			)
		}
	})
	if len(flagValueErrorAttributes) > 0 {
		return flagValueError.WithAttributes(flagValueErrorAttributes...)
	}
	return nil
}

var textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()

func setField(rv reflect.Value, path []string, v reflect.Value) error {
	rt := rv.Type()
	vt := v.Type()
	props := proto.GetProperties(rt)
	for _, prop := range props.Prop {
		if prop.OrigName == path[0] {
			field := rv.FieldByName(prop.Name)
			if field.Type().Kind() == reflect.Ptr {
				if field.IsNil() {
					field.Set(reflect.New(field.Type().Elem()))
				}
				field = field.Elem()
			}
			ft := field.Type()
			if len(path) == 1 {
				switch {
				case ft.AssignableTo(vt):
					field.Set(v)
				case vt.Kind() == reflect.String && reflect.PtrTo(ft).Implements(textUnmarshalerType):
					err := field.Addr().Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(v.String()))
					if err != nil {
						return err
					}
				case ft.PkgPath() == "time":
					switch {
					case ft.Name() == "Time" && vt.Kind() == reflect.String:
						var t time.Time
						var err error
						if v.String() != "" {
							t, err = time.Parse(time.RFC3339Nano, v.String())
							if err != nil {
								return err
							}
						}
						field.Set(reflect.ValueOf(t))
					case ft.Name() == "Duration" && vt.Kind() == reflect.String:
						d, err := time.ParseDuration(v.String())
						if err != nil {
							return err
						}
						field.Set(reflect.ValueOf(d))
					}
				case ft.PkgPath() == "github.com/gogo/protobuf/types":
					switch ft.Name() {
					case "DoubleValue":
						field.Set(reflect.ValueOf(types.DoubleValue{Value: v.Float()}))
					case "FloatValue":
						field.Set(reflect.ValueOf(types.FloatValue{Value: float32(v.Float())}))
					case "Int64Value":
						field.Set(reflect.ValueOf(types.Int64Value{Value: v.Int()}))
					case "UInt64Value":
						field.Set(reflect.ValueOf(types.UInt64Value{Value: v.Uint()}))
					case "Int32Value":
						field.Set(reflect.ValueOf(types.Int32Value{Value: int32(v.Int())}))
					case "UInt32Value":
						field.Set(reflect.ValueOf(types.UInt32Value{Value: uint32(v.Uint())}))
					case "BoolValue":
						field.Set(reflect.ValueOf(types.BoolValue{Value: v.Bool()}))
					case "StringValue":
						field.Set(reflect.ValueOf(types.StringValue{Value: v.String()}))
					case "BytesValue":
						s := strings.TrimPrefix(v.String(), "0x")
						buf, err := hex.DecodeString(s)
						if err != nil {
							return err
						}
						field.Set(reflect.ValueOf(types.BytesValue{Value: buf}))
					case "Timestamp":
						var t time.Time
						var err error
						if v.String() != "" {
							t, err = time.Parse(time.RFC3339Nano, v.String())
							if err != nil {
								return err
							}
						}
						field.Set(reflect.ValueOf(*ttnpb.ProtoTimePtr(t)))
					case "Duration":
						d := v.Interface().(time.Duration)
						field.Set(reflect.ValueOf(*ttnpb.ProtoDurationPtr(d)))
					}
				case ft.PkgPath() == "go.thethings.network/lorawan-stack/v3/pkg/ttnpb":
					switch typeName := ft.Name(); typeName {
					case "BoolValue":
						field.Set(reflect.ValueOf(ttnpb.BoolValue{Value: v.Bool()}))
					case "FrequencyValue":
						field.Set(reflect.ValueOf(ttnpb.FrequencyValue{Value: v.Uint()}))
					case "ZeroableFrequencyValue":
						field.Set(reflect.ValueOf(ttnpb.ZeroableFrequencyValue{Value: v.Uint()}))

					case "DataRateIndexValue":
						if enumValue, ok := proto.EnumValueMap(unwrapLoRaWANEnumType(typeName))[v.String()]; ok {
							field.Set(reflect.ValueOf(ttnpb.DataRateIndexValue{Value: ttnpb.DataRateIndex(enumValue)}))
							break
						}
						var enum ttnpb.DataRateIndex
						if err := enum.UnmarshalText([]byte(v.String())); err != nil {
							field.Set(reflect.ValueOf(ttnpb.DataRateIndexValue{Value: enum}))
							break
						}
						return fmt.Errorf(`invalid value "%s" for %s`, v.String(), typeName)
					case "DataRateOffsetValue":
						if enumValue, ok := proto.EnumValueMap(unwrapLoRaWANEnumType(typeName))[v.String()]; ok {
							field.Set(reflect.ValueOf(ttnpb.DataRateOffsetValue{Value: ttnpb.DataRateOffset(enumValue)}))
							break
						}
						var enum ttnpb.DataRateOffset
						if err := enum.UnmarshalText([]byte(v.String())); err != nil {
							field.Set(reflect.ValueOf(ttnpb.DataRateOffsetValue{Value: enum}))
							break
						}
						return fmt.Errorf(`invalid value "%s" for %s`, v.String(), typeName)
					case "PingSlotPeriodValue":
						if enumValue, ok := proto.EnumValueMap(unwrapLoRaWANEnumType(typeName))[v.String()]; ok {
							field.Set(reflect.ValueOf(ttnpb.PingSlotPeriodValue{Value: ttnpb.PingSlotPeriod(enumValue)}))
							break
						}
						var enum ttnpb.PingSlotPeriod
						if err := enum.UnmarshalText([]byte(v.String())); err != nil {
							field.Set(reflect.ValueOf(ttnpb.PingSlotPeriodValue{Value: enum}))
							break
						}
						return fmt.Errorf(`invalid value "%s" for %s`, v.String(), typeName)
					case "AggregatedDutyCycleValue":
						if enumValue, ok := proto.EnumValueMap(unwrapLoRaWANEnumType(typeName))[v.String()]; ok {
							field.Set(reflect.ValueOf(ttnpb.AggregatedDutyCycleValue{Value: ttnpb.AggregatedDutyCycle(enumValue)}))
							break
						}
						var enum ttnpb.AggregatedDutyCycle
						if err := enum.UnmarshalText([]byte(v.String())); err != nil {
							field.Set(reflect.ValueOf(ttnpb.AggregatedDutyCycleValue{Value: enum}))
							break
						}
						return fmt.Errorf(`invalid value "%s" for %s`, v.String(), typeName)
					case "RxDelayValue":
						if enumValue, ok := proto.EnumValueMap(unwrapLoRaWANEnumType(typeName))[v.String()]; ok {
							field.Set(reflect.ValueOf(ttnpb.RxDelayValue{Value: ttnpb.RxDelay(enumValue)}))
							break
						}
						var enum ttnpb.RxDelay
						if err := enum.UnmarshalText([]byte(v.String())); err != nil {
							field.Set(reflect.ValueOf(ttnpb.RxDelayValue{Value: enum}))
							break
						}
						return fmt.Errorf(`invalid value "%s" for %s`, v.String(), typeName)
					case "DeviceEIRPValue":
						if enumValue, ok := proto.EnumValueMap(unwrapLoRaWANEnumType(typeName))[v.String()]; ok {
							field.Set(reflect.ValueOf(ttnpb.DeviceEIRPValue{Value: ttnpb.DeviceEIRP(enumValue)}))
							break
						}
						var enum ttnpb.DeviceEIRP
						if err := enum.UnmarshalText([]byte(v.String())); err != nil {
							field.Set(reflect.ValueOf(ttnpb.DeviceEIRPValue{Value: enum}))
							break
						}
						return fmt.Errorf(`invalid value "%s" for %s`, v.String(), typeName)
					case "ADRAckDelayExponentValue":
						if enumValue, ok := proto.EnumValueMap(unwrapLoRaWANEnumType(typeName))[v.String()]; ok {
							field.Set(reflect.ValueOf(ttnpb.ADRAckDelayExponentValue{Value: ttnpb.ADRAckDelayExponent(enumValue)}))
							break
						}
						var enum ttnpb.ADRAckDelayExponent
						if err := enum.UnmarshalText([]byte(v.String())); err != nil {
							field.Set(reflect.ValueOf(ttnpb.ADRAckDelayExponentValue{Value: enum}))
							break
						}
						return fmt.Errorf(`invalid value "%s" for %s`, v.String(), typeName)
					case "ADRAckLimitExponentValue":
						if enumValue, ok := proto.EnumValueMap(unwrapLoRaWANEnumType(typeName))[v.String()]; ok {
							field.Set(reflect.ValueOf(ttnpb.ADRAckLimitExponentValue{Value: ttnpb.ADRAckLimitExponent(enumValue)}))
							break
						}
						var enum ttnpb.ADRAckLimitExponent
						if err := enum.UnmarshalText([]byte(v.String())); err != nil {
							field.Set(reflect.ValueOf(ttnpb.ADRAckLimitExponentValue{Value: enum}))
							break
						}
						return fmt.Errorf(`invalid value "%s" for %s`, v.String(), typeName)
					case "ADRSettings_DisabledMode":
						field.Set(reflect.ValueOf(ttnpb.ADRSettings_DisabledMode{}))
						break
					case "ADRSettings_StaticMode":
						field.Set(reflect.ValueOf(ttnpb.ADRSettings_StaticMode{}))
						break
					case "ADRSettings_DynamicMode":
						field.Set(reflect.ValueOf(ttnpb.ADRSettings_DynamicMode{}))
						break
					}
				case ft.Kind() == reflect.Slice && ft.Elem().Kind() == reflect.Uint8 && vt.Kind() == reflect.String:
					s := strings.TrimPrefix(v.String(), "0x")
					buf, err := hex.DecodeString(s)
					if err != nil {
						return err
					}
					field.Set(reflect.ValueOf(buf))
				case ft.Kind() == reflect.Array && ft.Elem().Kind() == reflect.Uint8 && vt.Kind() == reflect.String:
					s := strings.TrimPrefix(v.String(), "0x")
					buf, err := hex.DecodeString(s)
					if err != nil {
						return err
					}
					if len(buf) > 0 {
						if len(buf) != ft.Len() {
							return fmt.Errorf(`bytes of "%s" do not fit in [%d]byte`, v.String(), ft.Len())
						}
						for i := 0; i < ft.Len(); i++ {
							field.Index(i).Set(reflect.ValueOf(buf[i]))
						}
					} else {
						field.Set(reflect.Zero(ft))
					}
				case ft.Kind() == reflect.Slice && vt.Kind() == reflect.Slice:
					fte := ft.Elem()
					slice := reflect.MakeSlice(ft, v.Len(), v.Len())
					switch {
					case vt.Elem().ConvertibleTo(fte):
						for i := 0; i < v.Len(); i++ {
							slice.Index(i).Set(v.Index(i).Convert(fte))
						}
					case fte.Kind() == reflect.Ptr &&
						fte.Elem().PkgPath() == "go.thethings.network/lorawan-stack/v3/pkg/ttnpb" &&
						fte.Elem().Name() == "GatewayAntennaIdentifiers" && vt.Elem().Kind() == reflect.String:
						for i := 0; i < v.Len(); i++ {
							slice.Index(i).Set(reflect.ValueOf(&ttnpb.GatewayAntennaIdentifiers{
								GatewayIds: &ttnpb.GatewayIdentifiers{
									GatewayId: v.Index(i).String(),
								},
							}))
						}
					case vt.Elem().Kind() == reflect.String && reflect.PtrTo(fte).Implements(textUnmarshalerType):
						for i := 0; i < v.Len(); i++ {
							err := slice.Index(i).Addr().Interface().(encoding.TextUnmarshaler).UnmarshalText([]byte(v.Index(i).String()))
							if err != nil {
								return err
							}
						}
					default:
						return fmt.Errorf("%v is not convertible to %v", ft, vt)
					}
					field.Set(slice)
				default:
					return fmt.Errorf("%v is not assignable to %v", ft, vt)
				}
				return nil
			}
			for name, oneofProp := range props.OneofTypes {
				if name != path[1] {
					continue
				}
				if field.Type().Kind() != reflect.Interface {
					panic("oneof field is not an interface")
				}
				elem := field.Elem()
				switch {
				case !elem.IsValid():
					field.Set(reflect.New(oneofProp.Type.Elem()))
				case elem.Type() != oneofProp.Type:
					return fmt.Errorf("different oneof type")
				}
				field = field.Elem()
				if field.Type().Kind() == reflect.Ptr {
					if field.IsNil() {
						field.Set(reflect.New(field.Type().Elem()))
					}
					field = field.Elem()
				}
			}
			return setField(field, path[1:], v)
		}
	}
	return fmt.Errorf("unknown field")
}

// SelectAllFlagSet returns a flagset with the --all flag
func SelectAllFlagSet(what string) *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.Bool("all", false, fmt.Sprintf("select all %s fields", what))
	return flagSet
}

// UnsetFlagSet returns a flagset with the --unset flag
func UnsetFlagSet() *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.StringSlice("unset", []string{}, "list of fields to unset")
	return flagSet
}
