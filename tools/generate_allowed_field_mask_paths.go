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

//go:build ignore
// +build ignore

package main

import (
	"encoding/json"
	"os"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

var exceptions = make(map[string]func([]string) []string, 0)

func main() {
	messagesFile := "sdk/js/generated/allowed-field-mask-paths.json"
	if len(os.Args) == 2 {
		messagesFile = os.Args[1]
	}

	data, err := json.MarshalIndent(func() map[string][]string {
		ret := make(map[string][]string, len(ttnpb.RPCFieldMaskPaths))
		for rpc, v := range ttnpb.RPCFieldMaskPaths {
			f, ok := exceptions[rpc]
			if ok {
				ret[rpc] = f(v.Allowed)
			} else {
				ret[rpc] = v.Allowed
			}
		}
		return ret
	}(), "", "  ")
	if err != nil {
		panic(err)
	}

	if err = os.WriteFile(messagesFile, data, 0o644); err != nil {
		panic(err)
	}
}

func init() {
	// This prevents the console from accessing the JS for CAC related operations.
	// TODO: Remove this logic when CAC usage in the JS is removed.
	// (https://github.com/TheThingsNetwork/lorawan-stack/issues/5631)
	claimAuthenticationCodePaths := []string{
		"claim_authentication_code",
		"claim_authentication_code.value",
		"claim_authentication_code.valid_from",
		"claim_authentication_code.valid_to",
	}
	// Register exceptions.
	exceptions["/ttn.lorawan.v3.JsEndDeviceRegistry/Get"] = func(allowed []string) []string {
		return ttnpb.ExcludeFields(allowed, claimAuthenticationCodePaths...)
	}
	exceptions["/ttn.lorawan.v3.JsEndDeviceRegistry/Set"] = func(allowed []string) []string {
		return ttnpb.ExcludeFields(allowed, claimAuthenticationCodePaths...)
	}
}
