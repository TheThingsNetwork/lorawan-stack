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
	evtCreateApplication = events.Define("is.application.create", "create application")
	evtUpdateApplication = events.Define("is.application.update", "update application")
	evtDeleteApplication = events.Define("is.application.delete", "delete application")

	evtCreateClient = events.Define("is.client.create", "create OAuth client")
	evtUpdateClient = events.Define("is.client.update", "update OAuth client")
	evtDeleteClient = events.Define("is.client.delete", "delete OAuth client")

	evtCreateGateway = events.Define("is.gateway.create", "create gateway")
	evtUpdateGateway = events.Define("is.gateway.update", "update gateway")
	evtDeleteGateway = events.Define("is.gateway.delete", "delete gateway")

	evtCreateOrganization = events.Define("is.organization.create", "create organization")
	evtUpdateOrganization = events.Define("is.organization.update", "update organization")
	evtDeleteOrganization = events.Define("is.organization.delete", "delete organization")

	evtCreateUser = events.Define("is.user.create", "create user")
	evtUpdateUser = events.Define("is.user.update", "update user")
	evtDeleteUser = events.Define("is.user.delete", "delete user")
)
