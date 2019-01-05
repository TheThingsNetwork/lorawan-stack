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

package gogoproto

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	pbtypes "github.com/gogo/protobuf/types"
)

var nameTagRegex *regexp.Regexp

func init() {
	nameTagRegex = regexp.MustCompile("name\\=[a-zA-Z_0-9]+")
}

// GoFieldsPaths converts protobuf FieldMask paths to Go fields paths.
//
// This implementation does not support separation by ",", but only paths separated by ".".
func GoFieldsPaths(pb *pbtypes.FieldMask, v interface{}) []string {
	var newFields []string
	if pb == nil || len(pb.Paths) == 0 {
		return newFields
	}

	goFields := goFieldsFromProtoMasks(reflect.ValueOf(v))
	for _, field := range pb.Paths {
		goName, ok := goFields[field]
		if ok {
			newFields = append(newFields, goName)
		} else {
			newFields = append(newFields, field)
		}
	}

	return newFields
}

func goFieldsFromProtoMasks(v reflect.Value) map[string]string {
	if !v.IsValid() {
		return nil
	}

	fields := make(map[string]string)
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return fields
	}

	for i := 0; i < v.NumField(); i++ {
		tag := v.Type().Field(i).Tag.Get("protobuf")
		if tag == "" {
			continue
		}

		protoName := strings.TrimPrefix(nameTagRegex.FindString(tag), "name=")
		if protoName == "" {
			continue
		}
		goName := v.Type().Field(i).Name
		fields[protoName] = goName

		subFields := goFieldsFromProtoMasks(v.Field(i))
		if len(subFields) == 0 {
			continue
		}

		for k, v := range subFields {
			fields[fmt.Sprintf("%s.%s", protoName, k)] = fmt.Sprintf("%s%s%s", goName, ".", v)
		}
	}

	return fields
}
