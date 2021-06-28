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

import "context"

// ValidateContext wraps the generated validator with (optionally context-based) custom checks.
func (req *CryptoServicePayloadRequest) ValidateContext(context.Context) error {
	return req.ValidateFields()
}

// ValidateContext wraps the generated validator with (optionally context-based) custom checks.
func (req *JoinAcceptMICRequest) ValidateContext(context.Context) error {
	return req.ValidateFields()
}

// ValidateContext wraps the generated validator with (optionally context-based) custom checks.
func (req *DeriveSessionKeysRequest) ValidateContext(context.Context) error {
	return req.ValidateFields()
}

// ValidateContext wraps the generated validator with (optionally context-based) custom checks.
func (req *GetRootKeysRequest) ValidateContext(context.Context) error {
	return req.ValidateFields()
}

// ValidateContext wraps the generated validator with (optionally context-based) custom checks.
func (req *ProvisionEndDevicesRequest) ValidateContext(context.Context) error {
	return req.ValidateFields()
}

// ValidateContext wraps the generated validator with (optionally context-based) custom checks.
func (req *GetApplicationActivationSettingsRequest) ValidateContext(context.Context) error {
	return req.ValidateFields()
}

// ValidateContext wraps the generated validator with (optionally context-based) custom checks.
func (req *SetApplicationActivationSettingsRequest) ValidateContext(context.Context) error {
	if len(req.FieldMask.GetPaths()) == 0 {
		return req.ValidateFields()
	}
	return req.ValidateFields(append(FieldsWithPrefix("settings", req.FieldMask.GetPaths()...),
		"application_ids",
	)...)
}

// ValidateContext wraps the generated validator with (optionally context-based) custom checks.
func (req *DeleteApplicationActivationSettingsRequest) ValidateContext(context.Context) error {
	return req.ValidateFields()
}
