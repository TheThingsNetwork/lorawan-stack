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

import "go.thethings.network/lorawan-stack/pkg/events"

var (
	evtGatewayCreated = events.Define("is.gateway.create", "create gateway")
	evtGatewayUpdated = events.Define("is.gateway.update", "update gateway")
	evtGatewayDeleted = events.Define("is.gateway.delete", "delete gateway")

	evtUserCreated = events.Define("is.user.create", "create user")
	evtUserUpdated = events.Define("is.user.update", "update user")
	evtUserDeleted = events.Define("is.user.delete", "delete user")

	evtApplicationCreated = events.Define("is.application.create", "create application")
	evtApplicationUpdated = events.Define("is.application.update", "update application")
	evtApplicationDeleted = events.Define("is.application.delete", "delete application")
)
