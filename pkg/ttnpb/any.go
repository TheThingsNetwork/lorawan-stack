// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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
	proto "github.com/gogo/protobuf/proto"
	pbtypes "github.com/gogo/protobuf/types"
)

// MarshalAny wraps the MarshalAny func in the protobuf library.
func MarshalAny(pb proto.Message) (*pbtypes.Any, error) {
	return pbtypes.MarshalAny(pb)
}

// MustMarshalAny converts the proto message to an Any, or panics.
func MustMarshalAny(pb proto.Message) *pbtypes.Any {
	any, err := MarshalAny(pb)
	if err != nil {
		panic(err)
	}
	return any
}

// UnmarshalAny wraps the UnmarshalAny func in the protobuf library.
func UnmarshalAny(any *pbtypes.Any, pb proto.Message) error {
	return pbtypes.UnmarshalAny(any, pb)
}
