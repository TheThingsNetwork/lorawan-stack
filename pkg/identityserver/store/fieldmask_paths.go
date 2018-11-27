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

package store

const ( // TODO: should probably move this to ttnpb
	// please keep this sorted
	adminField                  = "admin"
	antennasField               = "antennas"
	attributesField             = "attributes"
	autoUpdateField             = "auto_update"
	brandIDField                = "version_ids.brand_id"
	contactInfoField            = "contact_info"
	descriptionField            = "description"
	downlinkPathConstraintField = "downlink_path_constraint"
	endorsedField               = "endorsed"
	enforceDutyCycleField       = "enforce_duty_cycle"
	firmwareVersionField        = "version_ids.firmware_version"
	frequencyPlanIDField        = "frequency_plan_id"
	gatewayServerAddressField   = "gateway_server_address"
	grantsField                 = "grants"
	hardwareVersionField        = "version_ids.hardware_version"
	locationPublicField         = "location_public"
	modelIDField                = "version_ids.model_id"
	nameField                   = "name"
	passwordField               = "password"
	passwordUpdatedAtField      = "password_updated_at"
	primaryEmailAddressField    = "primary_email_address"
	redirectURIsField           = "redirect_uris"
	requirePasswordUpdateField  = "require_password_update"
	rightsField                 = "rights"
	scheduleDownlinkLateField   = "schedule_downlink_late"
	secretField                 = "secret"
	skipAuthorizationField      = "skip_authorization"
	stateField                  = "state"
	statusPublicField           = "status_public"
	updateChannelField          = "update_channel"
)
