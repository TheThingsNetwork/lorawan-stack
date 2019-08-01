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

//+build ignore

package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func main() {
	messagesFile := "sdk/js/generated/allowed-field-masks.json"
	if len(os.Args) == 2 {
		messagesFile = os.Args[1]
	}

	data, err := json.MarshalIndent(ttnpb.AllowedFieldMaskPathsForRPC, "", "  ")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(messagesFile, data, 0644)
	if err != nil {
		panic(err)
	}
}
