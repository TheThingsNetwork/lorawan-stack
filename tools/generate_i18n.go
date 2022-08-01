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
	"fmt"
	"os"

	_ "go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/commands"
	_ "go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-stack/commands"
	"go.thethings.network/lorawan-stack/v3/pkg/i18n"
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

	global := i18n.CloneGlobal()
	global.Merge(messages)

	updated := global.Updated()
	if len(updated) > 0 {
		fmt.Println("Updated the following messages:")
		for _, updated := range updated {
			fmt.Printf(" - %s\n", updated)
		}
	}

	deleted := global.Cleanup()
	if len(deleted) > 0 {
		fmt.Println("Deleted the following messages:")
		for _, deleted := range deleted {
			fmt.Printf(" - %s\n", deleted)
		}
	}

	count := make(map[string]int)
	for _, msg := range global {
		for lang := range msg.Translations {
			count[lang] = count[lang] + 1
		}
	}

	for lang, count := range count {
		fmt.Printf("Language: %s\tTranslations: %5d\tMissing: %5d\n", lang, count, len(global)-count)
	}

	err = global.WriteFile(messagesFile)
	if err != nil {
		panic(err)
	}
}
