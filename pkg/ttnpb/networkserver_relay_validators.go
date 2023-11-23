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

package ttnpb

import "context"

// ValidateContext wraps the generated validator with (optionally context-based) custom checks.
func (req *UpdateRelayRequest) ValidateContext(context.Context) error {
	if len(req.FieldMask.GetPaths()) == 0 {
		return req.ValidateFields()
	}
	return req.ValidateFields(append(FieldsWithPrefix("configuration", req.FieldMask.GetPaths()...),
		"end_device_ids.application_ids",
		"end_device_ids.device_id",
	)...)
}

// ValidateContext wraps the generated validator with (optionally context-based) custom checks.
func (req *UpdateRelayUplinkForwardingRuleRequest) ValidateContext(context.Context) error {
	if len(req.FieldMask.GetPaths()) == 0 {
		return req.ValidateFields()
	}
	return req.ValidateFields(append(FieldsWithPrefix("uplink_forwarding_rule", req.FieldMask.GetPaths()...),
		"end_device_ids.application_ids",
		"end_device_ids.device_id",
		"index",
	)...)
}
