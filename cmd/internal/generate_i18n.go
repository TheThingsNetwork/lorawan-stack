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

//+build ignore

package main

import (
	"fmt"
	"os"

	_ "go.thethings.network/lorawan-stack/cmd/ttn-lw-application-server/commands"
	_ "go.thethings.network/lorawan-stack/cmd/ttn-lw-gateway-server/commands"
	_ "go.thethings.network/lorawan-stack/cmd/ttn-lw-identity-server/commands"
	_ "go.thethings.network/lorawan-stack/cmd/ttn-lw-join-server/commands"
	_ "go.thethings.network/lorawan-stack/cmd/ttn-lw-network-server/commands"
	_ "go.thethings.network/lorawan-stack/cmd/ttn-lw-stack/commands"
	"go.thethings.network/lorawan-stack/pkg/i18n"
)

func main() {
	messagesFile := "config/messages.json"
	if len(os.Args) == 2 {
		messagesFile = os.Args[1]
	}

	messages, err := i18n.ReadFile(messagesFile)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	i18n.Global.Merge(messages)

	updated := i18n.Global.Updated()
	if len(updated) > 0 {
		fmt.Println("Updated the following messages:")
		for _, updated := range updated {
			fmt.Printf(" - %s\n", updated)
		}
	}

	deleted := i18n.Global.Cleanup()
	if len(deleted) > 0 {
		fmt.Println("Deleted the following messages:")
		for _, deleted := range deleted {
			fmt.Printf(" - %s\n", deleted)
		}
	}

	count := make(map[string]int)
	for _, msg := range i18n.Global {
		for lang := range msg.Translations {
			count[lang] = count[lang] + 1
		}
	}

	for lang, count := range count {
		fmt.Printf("Language: %s\tTranslations: %5d\tMissing: %5d\n", lang, count, len(i18n.Global)-count)
	}

	err = i18n.Global.WriteFile(messagesFile)
	if err != nil {
		panic(err)
	}
}
