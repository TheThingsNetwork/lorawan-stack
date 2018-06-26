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
	"go.thethings.network/lorawan-stack/pkg/events"
)

var (
	evtCreateApplication         = events.Define("is.application.create", "create application")
	evtUpdateApplication         = events.Define("is.application.update", "update application")
	evtDeleteApplication         = events.Define("is.application.delete", "delete application")
	evtGenerateApplicationAPIKey = events.Define("is.application.api_key.generate", "generate application API key")
	evtUpdateApplicationAPIKey   = events.Define("is.application.api_key.update", "update application API key")
	evtDeleteApplicationAPIKey   = events.Define("is.application.api_key.delete", "delete application API key")

	evtCreateClient = events.Define("is.client.create", "create OAuth client")
	evtUpdateClient = events.Define("is.client.update", "update OAuth client")
	evtDeleteClient = events.Define("is.client.delete", "delete OAuth client")

	evtCreateGateway         = events.Define("is.gateway.create", "create gateway")
	evtUpdateGateway         = events.Define("is.gateway.update", "update gateway")
	evtDeleteGateway         = events.Define("is.gateway.delete", "delete gateway")
	evtGenerateGatewayAPIKey = events.Define("is.gateway.api_key.generate", "generate gateway API key")
	evtUpdateGatewayAPIKey   = events.Define("is.gateway.api_key.update", "update gateway API key")
	evtDeleteGatewayAPIKey   = events.Define("is.gateway.api_key.delete", "delete gateway API key")

	evtCreateOrganization         = events.Define("is.organization.create", "create organization")
	evtUpdateOrganization         = events.Define("is.organization.update", "update organization")
	evtDeleteOrganization         = events.Define("is.organization.delete", "delete organization")
	evtGenerateOrganizationAPIKey = events.Define("is.organization.api_key.generate", "generate organization API key")
	evtUpdateOrganizationAPIKey   = events.Define("is.organization.api_key.update", "update organization API key")
	evtDeleteOrganizationAPIKey   = events.Define("is.organization.api_key.delete", "delete organization API key")

	evtCreateUser         = events.Define("is.user.create", "create user")
	evtUpdateUser         = events.Define("is.user.update", "update user")
	evtDeleteUser         = events.Define("is.user.delete", "delete user")
	evtGenerateUserAPIKey = events.Define("is.user.api_key.generate", "generate user API key")
	evtUpdateUserAPIKey   = events.Define("is.user.api_key.update", "update user API key")
	evtDeleteUserAPIKey   = events.Define("is.user.api_key.delete", "delete user API key")
)
