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
	"fmt"

	"github.com/TheThingsNetwork/ttn/cmd/internal/shared"
	"github.com/TheThingsNetwork/ttn/pkg/identityserver"
)

// DefaultIdentityServerConfig is the default configuration for the IdentityServer.
var DefaultIdentityServerConfig = identityserver.Config{
	DatabaseURI:      "postgres://root@localhost:26257/is_development?sslmode=disable",
	PublicURL:        fmt.Sprintf("http://localhost:%s", shared.DefaultServiceBase.HTTP.Listen),
	OrganizationName: "The Things Network",
	DefaultSettings:  identityserver.DefaultSettings,
}
