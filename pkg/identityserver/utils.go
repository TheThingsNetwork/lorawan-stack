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

package identityserver

import (
	"github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func fieldMaskContains(fieldMask *types.FieldMask, search string) bool {
	for _, path := range fieldMask.Paths {
		if path == search {
			return true
		}
	}
	return false
}

func cleanContactInfo(info []*ttnpb.ContactInfo) {
	for _, info := range info {
		info.ValidatedAt = nil
	}
}
