// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/spf13/pflag"
)

func SelectFieldMask(cmdFlags *pflag.FlagSet, fieldMaskFlags ...*pflag.FlagSet) (paths []string) {
	cmdFlags.Visit(func(flag *pflag.Flag) {
		for _, fieldMaskFlags := range fieldMaskFlags {
			if b, err := fieldMaskFlags.GetBool(flag.Name); err == nil && b {
				paths = append(paths, flag.Name)
				return
			}
		}
	})
	return
}

func UpdateFieldMask(cmdFlags *pflag.FlagSet, fieldMaskFlags ...*pflag.FlagSet) (paths []string) {
	cmdFlags.Visit(func(flag *pflag.Flag) {
		for _, fieldMaskFlags := range fieldMaskFlags {
			if fieldMaskFlags.Lookup(flag.Name) != nil {
				paths = append(paths, flag.Name)
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

func isAtomicType(t reflect.Type, maskOnly bool) bool {
	switch t.Name() {
	case "Time", "Duration":
		return true
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
	case "attributes", "contact_info", "password_updated_at", "temporary_password_created_at",
		"temporary_password_expires_at", "antennas":
		return false
	}
	return true
}

func enumValues(t reflect.Type) map[string]int32 {
	if t.PkgPath() == "go.thethings.network/lorawan-stack/pkg/ttnpb" {
		return proto.EnumValueMap(fmt.Sprintf("ttn.lorawan.v3.%s", t.Name()))
	}
	return nil
}

func addField(fs *pflag.FlagSet, name string, t reflect.Type, maskOnly bool) {
	if maskOnly {
		if t.Kind() == reflect.Struct && !isAtomicType(t, maskOnly) {
			fs.Bool(name, false, fmt.Sprintf("select the %s field and all sub-fields", name))
		} else {
			fs.Bool(name, false, fmt.Sprintf("select the %s field", name))
		}
		return
	}
	if t.Kind() == reflect.Struct && !isAtomicType(t, maskOnly) {
		return
	}
	if t.PkgPath() == "" {
		switch t.Kind() {
		case reflect.Bool:
			fs.Bool(name, false, "")
		case reflect.String:
			fs.String(name, "", "")
		case reflect.Int32:
			fs.Int32(name, 0, "")
		case reflect.Int64:
			fs.Int64(name, 0, "")
		case reflect.Uint32:
			fs.Uint32(name, 0, "")
		case reflect.Uint64:
			fs.Uint64(name, 0, "")
		case reflect.Float32:
			fs.Float32(name, 0, "")
		case reflect.Float64:
			fs.Float64(name, 0, "")
		case reflect.Slice:
			switch t.Elem().Kind() {
			case reflect.String:
				fs.StringSlice(name, nil, "")
			case reflect.Int32:
				if valueMap := enumValues(t); valueMap != nil {
					values := make([]string, 0, len(valueMap))
					for value := range valueMap {
						values = append(values, value)
					}
					fs.StringSlice(name, nil, strings.Join(values, "|"))
				} else {
					fs.IntSlice(name, nil, "")
				}
			case reflect.Int64:
				fs.IntSlice(name, nil, "")
			case reflect.Uint8:
				fs.String(name, "", "(hex)")
			case reflect.Uint32, reflect.Uint64:
				fs.UintSlice(name, nil, "")
			case reflect.Ptr:
				// Not supported
			default:
				fmt.Printf("flags: %s slice not yet supported (%s)\n", t.Elem().Kind(), name)
			}
		case reflect.Map:
			// Not supported
		default:
			fmt.Printf("flags: %s not yet supported (%s)\n", t.Kind(), name)
		}
	} else if t.Kind() == reflect.Int32 && strings.HasSuffix(t.PkgPath(), "ttnpb") {
		if valueMap := enumValues(t); valueMap != nil {
			values := make([]string, 0, len(valueMap))
			for value := range valueMap {
				values = append(values, value)
			}
			fs.String(name, "", strings.Join(values, "|"))
		}
	} else if (t.Kind() == reflect.Slice || t.Kind() == reflect.Array) && t.Elem().Kind() == reflect.Uint8 {
		fs.String(name, "", "(hex)")
	} else {
		switch t.Name() {
		case "Time":
			fs.String(name, "", "(YYYY-MM-DDTHH:MM:SSZ)")
		case "Duration":
			fs.Duration(name, 0, "(1h2m3s)")
		default:
			fmt.Printf("flags: %s.%s not yet supported (%s)\n", t.PkgPath(), t.Name(), name)
		}
	}
}

func fieldMaskFlags(prefix []string, t reflect.Type, maskOnly bool) *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	props := proto.GetProperties(t)
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
		addField(flagSet, name, fieldType, maskOnly)
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

func SetFields(dst interface{}, flags *pflag.FlagSet, prefix ...string) {
	rv := reflect.Indirect(reflect.ValueOf(dst))
	flags.VisitAll(func(flag *pflag.Flag) {
		var v interface{}
		switch flag.Value.Type() {
		case "bool":
			v, _ = flags.GetBool(flag.Name)
		case "string":
			v, _ = flags.GetString(flag.Name)
		case "int32":
			v, _ = flags.GetInt32(flag.Name)
		case "int64":
			v, _ = flags.GetInt64(flag.Name)
		case "uint32":
			v, _ = flags.GetUint32(flag.Name)
		case "uint64":
			v, _ = flags.GetUint64(flag.Name)
		case "float32":
			v, _ = flags.GetFloat32(flag.Name)
		case "float64":
			v, _ = flags.GetFloat64(flag.Name)
		case "stringSlice":
			v, _ = flags.GetStringSlice(flag.Name)
		case "intSlice":
			v, _ = flags.GetIntSlice(flag.Name)
		case "uintSlice":
			v, _ = flags.GetUintSlice(flag.Name)
		case "duration":
			v, _ = flags.GetDuration(flag.Name)
		}
		if v == nil {
			panic(fmt.Sprintf("can't set %s to %s (%v)", flag.Name, flag.Value, flag.Value.Type()))
		}
		setField(rv, trimPrefix(strings.Split(flag.Name, "."), prefix...), reflect.ValueOf(v))
	})
}

func setField(rv reflect.Value, path []string, v reflect.Value) {
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
				case ft.Kind() == reflect.Int32 && vt.Kind() == reflect.String:
					if valueMap := enumValues(ft); valueMap != nil {
						field.SetInt(int64(valueMap[v.String()]))
					}
				case ft.PkgPath() == "time" && ft.Name() == "Time" && vt.Kind() == reflect.String:
					var t time.Time
					var err error
					if v.String() != "" {
						t, err = time.Parse(time.RFC3339Nano, v.String())
						if err != nil {
							panic(err)
						}
					}
					field.Set(reflect.ValueOf(t))
				case ft.PkgPath() == "time" && ft.Name() == "Duration" && vt.Kind() == reflect.String:
					d, err := time.ParseDuration(v.String())
					if err != nil {
						panic(err)
					}
					field.Set(reflect.ValueOf(d))
				case ft.Kind() == reflect.Slice && ft.Elem().Kind() == reflect.Uint8 && vt.Kind() == reflect.String:
					s := strings.TrimPrefix(v.String(), "0x")
					buf, err := hex.DecodeString(s)
					if err != nil {
						panic(err)
					}
					field.Set(reflect.ValueOf(buf))
				case ft.Kind() == reflect.Array && ft.Elem().Kind() == reflect.Uint8 && vt.Kind() == reflect.String:
					s := strings.TrimPrefix(v.String(), "0x")
					buf, err := hex.DecodeString(s)
					if err != nil {
						panic(err)
					}
					if len(buf) > 0 {
						if len(buf) != ft.Len() {
							panic(fmt.Errorf(`bytes of "%s" do not fit in [%d]byte`, v.String(), ft.Len()))
						}
						for i := 0; i < ft.Len(); i++ {
							field.Index(i).Set(reflect.ValueOf(buf[i]))
						}
					} else {
						field.Set(reflect.Zero(ft))
					}
				case ft.Kind() == reflect.Slice && vt.Kind() == reflect.Slice:
					if vt.Elem().ConvertibleTo(ft.Elem()) {
						slice := reflect.MakeSlice(ft, v.Len(), v.Len())
						for i := 0; i < v.Len(); i++ {
							slice.Index(i).Set(v.Index(i).Convert(ft.Elem()))
						}
						field.Set(slice)
					} else {
						panic(fmt.Sprintf("%v is not convertible to %v\n", ft, vt))
					}
				default:
					panic(fmt.Sprintf("%v is not assignable to %v\n", ft, vt))
				}
				return
			}
			setField(field, path[1:], v)
		}
	}
}
