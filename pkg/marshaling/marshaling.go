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

// Package marshaling implements marshaling, unmarshaling and diff functionality for structs and maps.
package marshaling

import (
	"reflect"

	"github.com/gogo/protobuf/proto"
)

var protoMessageType = reflect.TypeOf((*proto.Message)(nil)).Elem()

// Encoding represents the encoding used to encode value into []byte representation.
// This is used as the first byte in the encoded []byte representation and allows for consistent decoding.
type Encoding byte

const (
	// Separator is used to separate the flattened struct fields.
	Separator = "."

	// NOTE: The following list MUST NOT be reordered

	// RawEncoding represents case when value is encoded into "raw" byte value.
	RawEncoding Encoding = 1
	// JSONEncoding represents case when MarshalJSON() method was used to encode value.
	JSONEncoding Encoding = 2
	// ProtoEncoding represents case when Proto() method was used to encode value.
	ProtoEncoding Encoding = 3
	// GobEncoding represents case when Gob was used to encode value.
	GobEncoding Encoding = 4
	// MsgPackEncoding represents case when MsgPack was used to encode value.
	MsgPackEncoding Encoding = 5
)
