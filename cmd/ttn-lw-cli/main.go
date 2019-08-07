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

// ttn-lw-cli is the binary for the Command-line interface of The Things Stack for LoRaWAN.
package main

import (
	"os"

	cli_errors "go.thethings.network/lorawan-stack/cmd/internal/errors"
	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/commands"
	"go.thethings.network/lorawan-stack/pkg/errors"
)

func main() {
	if err := commands.Root.Execute(); err != nil {
		if errors.IsCanceled(err) {
			os.Exit(130)
		}
		cli_errors.PrintStack(os.Stderr, err)
		os.Exit(-1)
	}
}
