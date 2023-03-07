// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package alcsyncv1

import (
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/protobuf/types/known/structpb"
)

var errInvalidFieldType = errors.DefineCorruption("invalid_field_type", "field `{field}` has the wrong type `{type}`")

var defaultThreshold = time.Duration(4) * time.Second

type packageData struct {
	FPort     uint32
	Threshold time.Duration
}

func (d *packageData) fromStruct(st *structpb.Struct) error {
	fields := st.GetFields()
	value, ok := fields["f_port"]
	if ok {
		numberValue, ok := value.GetKind().(*structpb.Value_NumberValue)
		if !ok {
			return errInvalidFieldType.WithAttributes(
				"field", "f_port",
				"type", "number",
			)
		}
		d.FPort = uint32(numberValue.NumberValue)
	}

	value, ok = fields["threshold"]
	if ok {
		numberValue, ok := value.GetKind().(*structpb.Value_NumberValue)
		if !ok {
			return errInvalidFieldType.WithAttributes(
				"field", "threshold",
				"type", "number",
			)
		}
		d.Threshold = time.Duration(numberValue.NumberValue) * time.Second
	}
	return nil
}

func mergePackageData(
	def *ttnpb.ApplicationPackageDefaultAssociation,
	assoc *ttnpb.ApplicationPackageAssociation,
) (*packageData, error) {
	var defaultData, associationData packageData
	if err := defaultData.fromStruct(def.GetData()); err != nil {
		return nil, err
	}
	if err := associationData.fromStruct(assoc.GetData()); err != nil {
		return nil, err
	}

	merged := &packageData{}
	for _, data := range []packageData{defaultData, associationData} {
		if data.FPort != 0 {
			merged.FPort = data.FPort
		}
		if data.Threshold != 0 {
			merged.Threshold = data.Threshold
		}
	}
	if merged.Threshold == 0 {
		merged.Threshold = defaultThreshold
	}
	if merged.FPort == 0 {
		merged.FPort = def.Ids.FPort
	}
	return merged, nil
}
