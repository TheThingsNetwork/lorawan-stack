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
	evtGatewayCreated = events.Define("gateway.create", "create gateway")
	evtGatewayUpdated = events.Define("gateway.update", "update gateway")
	evtGatewayDeleted = events.Define("gateway.delete", "delete gateway")

	evtUserCreated = events.Define("user.create", "create user")
	evtUserUpdated = events.Define("user.update", "update user")
	evtUserDeleted = events.Define("user.delete", "delete user")

	evtApplicationCreated = events.Define("application.create", "create application")
	evtApplicationUpdated = events.Define("application.update", "update application")
	evtApplicationDeleted = events.Define("application.delete", "delete application")
)
