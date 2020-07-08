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

package registry

import "go.thethings.network/lorawan-stack/v3/pkg/ttnpb"

// ReplacedEndDeviceField provides how to transform an old End Device field to a new field.
type ReplacedEndDeviceField struct {
	Old          string
	New          string
	GetTransform func(dev *ttnpb.EndDevice)
	SetTransform func(dev *ttnpb.EndDevice, useOld, useNew bool) error
}

// ReplacedEndDeviceFieldMatch provides information about an End Device field replacement based on usage of the old name, new name or both.
type ReplacedEndDeviceFieldMatch struct {
	ReplacedEndDeviceField
	MatchedOld bool
	MatchedNew bool
}

// MatchReplacedEndDeviceFields returns a set of ReplacedEndDeviceFieldMatch based on usage by the given paths.
func MatchReplacedEndDeviceFields(paths []string, replaced []ReplacedEndDeviceField) ([]string, []ReplacedEndDeviceFieldMatch) {
	var matched []ReplacedEndDeviceFieldMatch
	for _, f := range replaced {
		hasOld, hasNew := ttnpb.HasAnyField(paths, f.Old), ttnpb.HasAnyField(paths, f.New)
		switch {
		case !hasOld && !hasNew:
			continue
		case hasOld && hasNew:
		case hasOld:
			paths = ttnpb.AddFields(paths, f.New)
		case hasNew:
			paths = ttnpb.AddFields(paths, f.Old)
		}
		matched = append(matched, ReplacedEndDeviceFieldMatch{
			ReplacedEndDeviceField: f,
			MatchedOld:             hasOld,
			MatchedNew:             hasNew,
		})
	}
	return paths, matched
}
