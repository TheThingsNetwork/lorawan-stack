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

package ttnpb

import "fmt"

// FieldIsZero returns whether path p is zero.
func (v *Picture_Embedded) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "data":
		return v.Data == nil
	case "mime_type":
		return v.MimeType == ""
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}

// FieldIsZero returns whether path p is zero.
func (v *Picture) FieldIsZero(p string) bool {
	if v == nil {
		return true
	}
	switch p {
	case "embedded":
		return v.Embedded == nil
	case "embedded.data":
		return v.Embedded.FieldIsZero("embedded.data")
	case "embedded.mime_type":
		return v.Embedded.FieldIsZero("embedded.mime_type")
	case "sizes":
		return v.Sizes == nil
	}
	panic(fmt.Sprintf("unknown path '%s'", p))
}
