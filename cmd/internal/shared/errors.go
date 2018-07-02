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

package shared

import errors "go.thethings.network/lorawan-stack/pkg/errorsv3"

// Errors returned by component initialization.
var (
	ErrBaseComponentInitialize     = errors.Define("base_component_initialize", "could not initialize base component")
	ErrIdentityServerInitialize    = errors.Define("identity_server_initialize", "could not initialize identity server")
	ErrGatewayServerInitialize     = errors.Define("gateway_server_initialize", "could not initialize gateway server")
	ErrNetworkServerInitialize     = errors.Define("network_server_initialize", "could not initialize network server")
	ErrApplicationServerInitialize = errors.Define("application_server_initialize", "could not initialize application server")
	ErrJoinServerInitialize        = errors.Define("join_server_initialize", "could not initialize join server")
	ErrConsoleInitialize           = errors.Define("console_initialize", "could not initialize console")
)
