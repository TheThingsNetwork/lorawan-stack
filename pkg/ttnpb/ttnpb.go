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

package ttnpb

import (
	"encoding/gob"

	pbtypes "github.com/gogo/protobuf/types"
)

func init() {
	gob.Register(&pbtypes.Value_NullValue{})
	gob.Register(&pbtypes.Value_NumberValue{})
	gob.Register(&pbtypes.Value_StringValue{})
	gob.Register(&pbtypes.Value_BoolValue{})
	gob.Register(&pbtypes.Value_StructValue{})
	gob.Register(&pbtypes.Value_ListValue{})
	gob.Register(&pbtypes.Struct{})
	gob.Register(&pbtypes.Value{})
}
