// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

// ttn-lw-identity-server is the binary that runs the Identity Server of The Things Network Stack for LoRaWAN.
package main

import (
	"fmt"
	"os"
	"strings"

	"go.thethings.network/lorawan-stack/cmd/ttn-lw-identity-server/commands"
	"go.thethings.network/lorawan-stack/pkg/errors"
)

func main() {
	if err := commands.Root.Execute(); err != nil {
		for i, err := range errors.Stack(err) {
			fmt.Println(strings.Repeat("  ", i), err)
		}
		os.Exit(-1)
	}
}
