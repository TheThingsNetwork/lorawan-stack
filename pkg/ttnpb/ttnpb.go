// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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
