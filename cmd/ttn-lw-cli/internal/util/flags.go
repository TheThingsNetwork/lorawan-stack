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
	"fmt"
	"reflect"
	"strings"

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

func FieldFlags(v interface{}) *pflag.FlagSet {
	t := reflect.Indirect(reflect.ValueOf(v)).Type()
	return fieldMaskFlags(nil, t, false)
}

func FieldMaskFlags(v interface{}) *pflag.FlagSet {
	t := reflect.Indirect(reflect.ValueOf(v)).Type()
	return fieldMaskFlags(nil, t, true)
}

func isAtomicType(t reflect.Type) bool {
	switch t.Name() {
	case "Time":
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
		if t.Kind() == reflect.Struct && !isAtomicType(t) {
			fs.Bool(name, false, fmt.Sprintf("select the %s field and all sub-fields", name))
		} else {
			fs.Bool(name, false, fmt.Sprintf("select the %s field", name))
		}
		return
	}
	if t.Kind() == reflect.Struct && !isAtomicType(t) {
		return
	}
	if t.PkgPath() == "" {
		switch t.Kind() {
		case reflect.Bool:
			fs.Bool(name, false, "")
		case reflect.String:
			fs.String(name, "", "")
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
				}
			default:
				fmt.Printf("flags: %s slice not yet supported (%s)\n", t.Elem().Kind(), name)
			}
		default:
			fmt.Printf("flags: %s not yet supported (%s)\n", t.Kind(), name)
		}
	} else {
		switch t.Kind() {
		case reflect.Int32:
			if valueMap := enumValues(t); valueMap != nil {
				values := make([]string, 0, len(valueMap))
				for value := range valueMap {
					values = append(values, value)
				}
				fs.String(name, "", strings.Join(values, "|"))
			}
		default:
			fmt.Printf("flags: %s.%s not yet supported (%s)\n", t.PkgPath(), t.Name(), name)
		}
	}
}

func fieldMaskFlags(prefix []string, t reflect.Type, maskOnly bool) *pflag.FlagSet {
	flagSet := new(pflag.FlagSet)
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
		if fieldType.Kind() == reflect.Struct {
			flagSet.AddFlagSet(fieldMaskFlags(path, fieldType, maskOnly))
		}
	}
	return flagSet
}

func SetFields(dst interface{}, flags *pflag.FlagSet) {
	rv := reflect.Indirect(reflect.ValueOf(dst))
	flags.VisitAll(func(flag *pflag.Flag) {
		var v interface{}
		switch flag.Value.Type() {
		case "bool":
			v, _ = flags.GetBool(flag.Name)
			if v == "" {
				return
			}
		case "string":
			v, _ = flags.GetString(flag.Name)
			if v == "" {
				return
			}
		case "stringSlice":
			s, _ := flags.GetStringSlice(flag.Name)
			if len(s) == 0 {
				return
			}
			v = s
		}
		if v == nil {
			panic(fmt.Sprintf("can't set %s to %s (%v)", flag.Name, flag.Value, flag.Value.Type()))
		}
		setField(rv, strings.Split(flag.Name, "."), reflect.ValueOf(v))
	})
}

func setField(rv reflect.Value, path []string, v reflect.Value) {
	rt := rv.Type()
	vt := v.Type()
	props := proto.GetProperties(rt)
	for _, prop := range props.Prop {
		if prop.OrigName == path[0] {
			field := rv.FieldByName(prop.Name)
			ft := field.Type()
			if len(path) == 1 {
				switch {
				case ft.AssignableTo(vt):
					field.Set(v)
				case ft.Kind() == reflect.Int32 && vt.Kind() == reflect.String:
					if valueMap := enumValues(ft); valueMap != nil {
						field.SetInt(int64(valueMap[v.String()]))
					}
				default:
					panic(fmt.Sprintf("%v is not assingable to %v\n", ft, vt))
				}
				return
			}
			if ft.Kind() == reflect.Ptr {
				if field.IsNil() {
					field.Set(reflect.New(ft.Elem()))
				}
				field = field.Elem()
			}
			setField(field, path[1:], v)
		}
	}
}
