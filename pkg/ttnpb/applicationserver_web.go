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

// IsZero reports whether ids represent zero identifiers.
func (ids *ApplicationWebhookIdentifiers) IsZero() bool {
	if ids == nil {
		return true
	}
	return ids.WebhookId == "" && ids.ApplicationIds.IsZero()
}

// All EntityType methods implement the IDStringer interface.

func (m *ApplicationWebhookIdentifiers) EntityType() string {
	return m.GetApplicationIds().EntityType()
}

func (m *ApplicationWebhook) EntityType() string {
	return m.GetIds().EntityType()
}

func (m *GetApplicationWebhookRequest) EntityType() string {
	return m.GetIds().EntityType()
}

func (m *ListApplicationWebhooksRequest) EntityType() string {
	return m.GetApplicationIds().EntityType()
}

func (m *SetApplicationWebhookRequest) EntityType() string {
	return m.GetWebhook().EntityType()
}

// All IDString methods implement the IDStringer interface.

func (m *ApplicationWebhookIdentifiers) IDString() string {
	return m.GetApplicationIds().IDString()
}

func (m *ApplicationWebhook) IDString() string {
	return m.GetIds().IDString()
}

func (m *GetApplicationWebhookRequest) IDString() string {
	return m.GetIds().IDString()
}

func (m *ListApplicationWebhooksRequest) IDString() string {
	return m.GetApplicationIds().IDString()
}

func (m *SetApplicationWebhookRequest) IDString() string {
	return m.GetWebhook().IDString()
}

// All ExtractRequestFields methods are used by github.com/grpc-ecosystem/go-grpc-middleware/tags.

func (m *ApplicationWebhookIdentifiers) ExtractRequestFields(dst map[string]interface{}) {
	m.GetApplicationIds().ExtractRequestFields(dst)
}

func (m *ApplicationWebhook) ExtractRequestFields(dst map[string]interface{}) {
	m.GetIds().ExtractRequestFields(dst)
}

func (m *GetApplicationWebhookRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetIds().ExtractRequestFields(dst)
}

func (m *ListApplicationWebhooksRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetApplicationIds().ExtractRequestFields(dst)
}

func (m *SetApplicationWebhookRequest) ExtractRequestFields(dst map[string]interface{}) {
	m.GetWebhook().ExtractRequestFields(dst)
}
