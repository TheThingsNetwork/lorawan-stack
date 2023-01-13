// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package assertions

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/google/go-cmp/cmp"
)

func getProtoMessage(i interface{}) proto.Message {
	v := reflect.ValueOf(i)
	switch v.Type().Kind() {
	case reflect.Struct:
		p := reflect.New(v.Type())
		p.Elem().Set(v)
		i = p.Interface()
	case reflect.Ptr:
		if v.IsNil() {
			return nil
		}
	default:
		return nil
	}
	if m, ok := i.(proto.Message); ok {
		return m
	}
	return nil
}

const shouldHaveEmptyProtoDiff = "Expected:\n%v\nActual:\n%v\nDiff:\n%s\n(should be empty diff)!"

// indentBlock indents a block of text with an indent string.
func indentBlock(text, indent string) string {
	if text == "" {
		return indent
	}
	if text[len(text)-1:] == "\n" {
		result := ""
		for _, j := range strings.Split(text[:len(text)-1], "\n") {
			result += indent + j + "\n"
		}
		return result
	}
	result := ""
	for _, j := range strings.Split(strings.TrimRight(text, "\n"), "\n") {
		result += indent + j + "\n"
	}
	return result[:len(result)-1]
}

func shouldEqualProto(actual, expected proto.Message) string {
	var actualBuf bytes.Buffer
	err := proto.MarshalText(&actualBuf, actual)
	if err != nil {
		return fmt.Sprintf("can't marshal actual proto for equality check: %v", err)
	}
	var expectedBuf bytes.Buffer
	err = proto.MarshalText(&expectedBuf, expected)
	if err != nil {
		return fmt.Sprintf("can't marshal expected proto for equality check: %v", err)
	}
	expectedText, actualText := expectedBuf.String(), actualBuf.String()
	if diff := cmp.Diff(actualText, expectedText); len(diff) > 0 {
		return fmt.Sprintf(shouldHaveEmptyProtoDiff, indentBlock(expectedText, "    "), indentBlock(actualText, "    "), diff)
	}
	return success
}
