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

package ttnpb

// EntityType implements the IDStringer interface.
func (m *SearchEndDevicesRequest) EntityType() string {
	return m.GetApplicationIds().EntityType()
}

// IDString implements the IDStringer interface.
func (m *SearchEndDevicesRequest) IDString() string {
	return m.GetApplicationIds().IDString()
}

// ExtractRequestFields is used by github.com/grpc-ecosystem/go-grpc-middleware/tags.
func (m *SearchEndDevicesRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetApplicationIds().ExtractRequestFields(dst)
}
