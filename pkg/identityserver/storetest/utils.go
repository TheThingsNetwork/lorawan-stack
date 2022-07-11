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

package storetest

import (
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func fieldMask(paths ...string) store.FieldMask {
	return ttnpb.ExcludeFields(
		paths,
		"ids",
		"created_at",
		"updated_at",
		"deleted_at",
	)
}

var attributes = map[string]string{
	"foo": "bar",
	"bar": "baz",
	"baz": "qux",
}

var updatedAttributes = map[string]string{
	"attribute": "new",
	"foo":       "bar",
	"bar":       "updated",
}
