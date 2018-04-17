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

package console

import (
	"fmt"

	"go.thethings.network/lorawan-stack/cmd/internal/shared"
	"go.thethings.network/lorawan-stack/pkg/console"
)

// DefaultConsoleConfig is the default configuration for the Console.
var DefaultConsoleConfig = console.Config{
	DefaultLanguage:   "en",
	IdentityServerURL: fmt.Sprintf("http://localhost%s/id", shared.DefaultServiceBase.HTTP.Listen),
	PublicURL:         fmt.Sprintf("http://localhost%s", shared.DefaultServiceBase.HTTP.Listen),
	OAuth: console.OAuth{
		ID:     "ttn-console",
		Secret: "ttn-console",
	},
}
