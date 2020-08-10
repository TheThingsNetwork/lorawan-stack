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
	"log"
	"os"

	jsoniter "github.com/json-iterator/go"
	_ "go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/commands"
	_ "go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-stack/commands"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
)

var streamPool jsoniter.StreamPool = jsoniter.Config{
	IndentionStep: 2,
}.Froze()

var json = jsonpb.TTN()

func main() {
	messagesFile := "doc/data/events.json"
	if len(os.Args) == 2 {
		messagesFile = os.Args[1]
	}
	f, err := os.OpenFile(messagesFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	enc := streamPool.BorrowStream(f)
	defer func() {
		enc.Flush()
		streamPool.ReturnStream(enc)
	}()
	enc.WriteObjectStart()
	for i, e := range events.All().Definitions() {
		if i > 0 {
			enc.WriteMore()
		}
		enc.WriteObjectField(e.Name())
		enc.WriteObjectStart()

		enc.WriteObjectField("name")
		enc.WriteString(e.Name())

		enc.WriteMore()

		enc.WriteObjectField("description")
		enc.WriteString(e.Description())

		if e.DataType() != nil {
			enc.WriteMore()

			enc.WriteObjectField("data")

			raw, err := json.Marshal(e.DataType())
			if err != nil {
				log.Fatal(err)
			}
			enc.WriteVal(jsoniter.RawMessage(raw))
		}

		enc.WriteObjectEnd()
	}
	enc.WriteObjectEnd()
}
