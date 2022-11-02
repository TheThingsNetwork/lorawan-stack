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

package loraclouddevicemanagementv1

import (
	"fmt"
	"net/url"

	"github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

type packageData struct {
	token          string
	serverURL      *url.URL
	fPortSet       map[uint32]struct{}
	useTLVEncoding *bool
}

const (
	tokenField          = "token"
	serverURLField      = "server_url"
	fPortSetField       = "f_port_set"
	useTLVEncodingField = "use_tlv_encoding"
)

func (d *packageData) GetUseTLVEncoding() bool {
	if d == nil || d.useTLVEncoding == nil {
		return false
	}
	return *d.useTLVEncoding
}

var (
	errFieldNotFound    = errors.DefineNotFound("field_not_found", "field `{field}` not found")
	errInvalidFieldType = errors.DefineCorruption("invalid_field_type", "field `{field}` has the wrong type `{type}`")
)

func (d *packageData) fromStruct(st *types.Struct) (err error) {
	fields := st.GetFields()
	value, ok := fields[tokenField]
	if !ok {
		return errFieldNotFound.WithAttributes("field", tokenField)
	}
	stringValue, ok := value.GetKind().(*types.Value_StringValue)
	if !ok {
		return errInvalidFieldType.WithAttributes(
			"field", tokenField,
			"type", fmt.Sprintf("%T", value.GetKind()),
		)
	}
	d.token = stringValue.StringValue
	value, ok = fields[serverURLField]
	if ok {
		stringValue, ok := value.GetKind().(*types.Value_StringValue)
		if !ok {
			return errInvalidFieldType.WithAttributes(
				"field", serverURLField,
				"type", fmt.Sprintf("%T", value.GetKind()),
			)
		}
		if d.serverURL, err = url.Parse(stringValue.StringValue); err != nil {
			return err
		}
	}
	value, ok = fields[useTLVEncodingField]
	if ok {
		boolValue, ok := value.GetKind().(*types.Value_BoolValue)
		if !ok {
			return errInvalidFieldType.WithAttributes(
				"field", useTLVEncodingField,
				"type", fmt.Sprintf("%T", value.GetKind()),
			)
		}
		d.useTLVEncoding = &boolValue.BoolValue
	}
	value, ok = fields[fPortSetField]
	if ok {
		listValue, ok := value.GetKind().(*types.Value_ListValue)
		if !ok {
			return errInvalidFieldType.WithAttributes(
				"field", fPortSetField,
				"type", fmt.Sprintf("%T", value.GetKind()),
			)
		}
		listValues := listValue.ListValue.GetValues()
		d.fPortSet = make(map[uint32]struct{}, len(listValues))
		for _, v := range listValues {
			numberValue, ok := v.GetKind().(*types.Value_NumberValue)
			if !ok {
				return errInvalidFieldType.WithAttributes(
					"field", fPortSetField,
					"type", fmt.Sprintf("%T", v.GetKind()),
				)
			}
			d.fPortSet[uint32(numberValue.NumberValue)] = struct{}{}
		}
	}
	return nil
}
