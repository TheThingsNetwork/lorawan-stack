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

package store

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

var nameTagRegex *regexp.Regexp

func init() {
	nameTagRegex = regexp.MustCompile("name\\=[a-zA-Z_0-9\\-]*")
}

func ConvertProtoFields(fields []string, st reflect.Value) []string {
	for st.Kind() == reflect.Ptr {
		st = st.Elem()
	}
	if st.Kind() != reflect.Struct {
		return fields
	}

	for i := 0; i < st.NumField(); i++ {
		if protoTag := st.Type().Field(i).Tag.Get("protobuf"); protoTag != "" {
			protoNames := nameTagRegex.FindAllString(protoTag, 1)
			if len(protoNames) == 0 {
				continue
			}

			newFields := []string{}

			protoName := strings.TrimRight(strings.TrimPrefix(protoNames[0], "name="), ",")
			protoNameAsPrefix := fmt.Sprintf("%s%s", protoName, Separator)
			GoName := st.Type().Field(i).Name

			trimmedFields := []string{}
			var fieldItselfPresent bool
			for _, field := range fields {
				if field == protoName {
					fieldItselfPresent = true
				} else if strings.HasPrefix(field, protoNameAsPrefix) {
					trimmedFields = append(trimmedFields, strings.TrimPrefix(field, protoNameAsPrefix))
				} else {
					newFields = append(newFields, field)
				}
			}

			trimmedConvertedFields := ConvertProtoFields(trimmedFields, st.Field(i))

			for _, trimmedConvertedField := range trimmedConvertedFields {
				fieldWithPrefix := strings.Join([]string{GoName, trimmedConvertedField}, Separator)
				newFields = append(newFields, fieldWithPrefix)
			}
			if fieldItselfPresent {
				newFields = append(newFields, GoName)
			}

			fields = newFields
		}
	}

	return fields
}
