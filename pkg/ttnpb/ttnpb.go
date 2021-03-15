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

package ttnpb

import (
	"strconv"
)

func unmarshalJSONString(b []byte) ([]byte, bool) {
	if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
		return b[1 : len(b)-1], true
	}
	return b, false
}

func unmarshalEnumFromBinary(typName string, names map[int32]string, b []byte) (int32, error) {
	if len(b) != 1 {
		return 0, errCouldNotParse(typName)(string(b))
	}
	i := int32(b[0])
	if _, ok := names[i]; !ok {
		return 0, errCouldNotParse(typName)(string(b))
	}
	return i, nil
}

func unmarshalEnumFromNumber(typName string, names map[int32]string, b []byte) (int32, error) {
	s := string(b)
	i64, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, errCouldNotParse(typName)(s).WithCause(err)
	}
	i := int32(i64)
	if _, ok := names[i]; !ok {
		return 0, errCouldNotParse(typName)(s)
	}
	return i, nil
}

func unmarshalEnumFromText(typName string, values map[string]int32, b []byte) (int32, error) {
	s := string(b)
	i, ok := values[s]
	if !ok {
		return 0, errCouldNotParse(typName)(s)
	}
	return i, nil
}

func unmarshalEnumFromTextPrefix(typName string, values map[string]int32, prefix string, b []byte) (int32, error) {
	s := string(b)
	i, ok := values[s]
	if ok {
		return i, nil
	}
	i, ok = values[prefix+s]
	if ok {
		return i, nil
	}
	return 0, errCouldNotParse(typName)(s)
}
