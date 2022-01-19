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

// All EntityType methods implement the IDStringer interface.

func (m *GetApplicationLinkRequest) EntityType() string {
	return m.GetApplicationIds().EntityType()
}

func (m *SetApplicationLinkRequest) EntityType() string {
	return m.GetApplicationIds().EntityType()
}

// All IDString methods implement the IDStringer interface.

func (m *GetApplicationLinkRequest) IDString() string {
	return m.GetApplicationIds().IDString()
}

func (m *SetApplicationLinkRequest) IDString() string {
	return m.GetApplicationIds().IDString()
}

// All ExtractRequestFields methods are used by github.com/grpc-ecosystem/go-grpc-middleware/tags.

func (m *GetApplicationLinkRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetApplicationIds().ExtractRequestFields(dst)
}

func (m *SetApplicationLinkRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetApplicationIds().ExtractRequestFields(dst)
}
