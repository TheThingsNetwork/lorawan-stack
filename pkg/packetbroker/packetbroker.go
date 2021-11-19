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

package packetbroker

import (
	"fmt"
	"time"
)

// Default values for Packet Broker IAM.
const (
	DefaultTokenIssuer       = "https://iam.packetbroker.net"
	DefaultTokenURL          = DefaultTokenIssuer + "/token"
	DefaultPublicKeyCacheTTL = 10 * time.Minute
)

// TokenPublicKeysURL returns the URL with public keys with which a token are signed.
func TokenPublicKeysURL(issuer string) string {
	return fmt.Sprintf("%s/.well-known/jwks.json", issuer)
}
