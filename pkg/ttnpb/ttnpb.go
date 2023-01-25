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
	"math"

	"go.thethings.network/lorawan-stack/v3/pkg/util/byteutil"
)

func marshalBinaryEnum(v int32) []byte {
	switch v := uint32(v); {
	case v <= 255:
		return []byte{byte(v)}
	case v <= math.MaxUint16:
		return byteutil.AppendUint16(make([]byte, 2), uint16(v), 2)
	case v <= byteutil.MaxUint24:
		return byteutil.AppendUint32(make([]byte, 3), v, 3)
	default:
		return byteutil.AppendUint32(make([]byte, 4), v, 4)
	}
}

func unmarshalEnumFromBinary(typName string, names map[int32]string, b []byte) (int32, error) {
	if len(b) > 4 {
		return 0, errCouldNotParse(typName)(string(b))
	}
	i := int32(byteutil.ParseUint32(b))
	if _, ok := names[i]; !ok {
		return 0, errCouldNotParse(typName)(string(b))
	}
	return i, nil
}
