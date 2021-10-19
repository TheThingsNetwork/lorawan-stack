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

//go:build ignore
// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path"

	_ "go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/commands"
	_ "go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-stack/commands"
	"go.thethings.network/lorawan-stack/v3/pkg/i18n"
	"golang.org/x/text/language"
)

var defaultLanguage = language.English

var languages = []language.Tag{
	language.Japanese,
}

func readLanguageFile(lang language.Tag, messages *i18n.Messages) error {
	dataFileName := path.Join("pkg", "locales", fmt.Sprintf("%s.json", lang))
	langData, err := os.ReadFile(dataFileName)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		langData = []byte{'{', '}'}
	}
	err = json.Unmarshal(langData, messages)
	if err != nil {
		return err
	}
	return nil
}

func writeLanguageFile(lang language.Tag, messages *i18n.Messages) error {
	dataFileName := path.Join("pkg", "locales", fmt.Sprintf("%s.json", lang))
	langData, err := json.MarshalIndent(messages, "", "  ")
	if err != nil {
		return err
	}
	err = os.WriteFile(dataFileName, langData, 0644)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	messages, err := i18n.Default().MessagesFor(defaultLanguage, false)
	if err != nil {
		log.Fatal(err)
	}
	messageIDs := messages.GetAllIDs()

	oldMessages, err := i18n.NewMessages(defaultLanguage)
	if err != nil {
		log.Fatal(err)
	}
	err = readLanguageFile(defaultLanguage, oldMessages)
	if err != nil {
		log.Fatal(err)
	}

	oldIDs, sameIDs, newIDs := diffSortedStrings(oldMessages.GetAllIDs(), messageIDs)
	for _, id := range newIDs {
		log.Printf("[**] new message %s", id)
	}
	changedIDs := make([]string, 0, len(sameIDs))
	for _, id := range sameIDs {
		if oldMessages.Get(id).String() != messages.Get(id).String() {
			log.Printf("[**] updated message %s", id)
			changedIDs = append(changedIDs, id)
		}
	}
	for _, id := range oldIDs {
		log.Printf("[**] removed message %s", id)
	}

	err = writeLanguageFile(defaultLanguage, messages)
	if err != nil {
		log.Fatal(err)
	}

	for _, lang := range languages {
		translations, err := i18n.Default().MessagesFor(lang, false)
		if err != nil {
			log.Fatal(err)
		}
		err = readLanguageFile(lang, translations)
		if err != nil {
			log.Fatal(err)
		}
		oldIDs, _, newIDs := diffSortedStrings(translations.GetAllIDs(), messageIDs)
		for _, id := range oldIDs {
			log.Printf("[%s] deleted message %s", lang, id)
			translations.Delete(id)
		}
		for _, id := range changedIDs {
			log.Printf("[%s] updated message %s needs translation", lang, id)
			translations.Delete(id)
			translations.Define(id, messages.Get(id).String()+fmt.Sprintf(" [translate %s-%s]", defaultLanguage, lang))
		}
		for _, id := range newIDs {
			log.Printf("[%s] new message %s (%v) needs translation", lang, id, messages.Get(id))
			translations.Define(id, messages.Get(id).String()+fmt.Sprintf(" [translate %s-%s]", defaultLanguage, lang))
		}
		err = writeLanguageFile(lang, translations)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func diffSortedStrings(a, b []string) (onlyInA, inBoth, onlyInB []string) {
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if a[i] < b[j] {
			onlyInA = append(onlyInA, a[i])
			i++
			continue
		}
		if a[i] > b[j] {
			onlyInB = append(onlyInB, b[j])
			j++
			continue
		}
		inBoth = append(inBoth, a[i])
		i++
		j++
	}
	if i < len(a) {
		onlyInA = append(onlyInA, a[i:]...)
	}
	if j < len(b) {
		onlyInB = append(onlyInB, b[j:]...)
	}
	return
}
