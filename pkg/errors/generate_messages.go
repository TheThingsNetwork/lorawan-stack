// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"

	"github.com/TheThingsNetwork/ttn/pkg/config"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/log"
)

const defaultLanguage = "en"

type cfg struct {
	Filename string `name:"filename" description:"the location of the file that contains the messages"`
}

type message struct {
	Key          string                `json:"key"`
	Descriptor   *errors.ErrDescriptor `json:"descriptor"`
	Translations map[string]string     `json:"translations"`
	used         bool                  `json:"-"`
}

func read(filename string) (map[string]*message, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}

		return nil, err
	}

	msgs := make([]*message, 0)
	err = json.Unmarshal(content, &msgs)
	if err != nil {
		return nil, err
	}

	res := make(map[string]*message)
	for _, msg := range msgs {
		msg.Translations[defaultLanguage] = msg.Descriptor.MessageFormat
		res[msg.Key] = msg
	}

	return res, nil
}

func key(namespace string, code errors.Code) string {
	return fmt.Sprintf("%s:%s", namespace, code)
}

func merge(old map[string]*message, new []*errors.ErrDescriptor) ([]*message, map[string]int) {
	stats := map[string]int{
		defaultLanguage: 0,
	}

	res := make([]*message, 0)
	for _, desc := range new {
		k := key(desc.Namespace, desc.Code)

		msg := old[k]
		if msg == nil {
			msg = &message{
				Key:        k,
				Descriptor: desc,
			}
		}

		msg.Descriptor = desc

		// clear translations if message changed
		if msg.Translations[defaultLanguage] != desc.MessageFormat {
			msg.Translations = map[string]string{
				defaultLanguage: desc.MessageFormat,
			}
		}

		res = append(res, msg)
		for lang, _ := range msg.Translations {
			stats[lang] = stats[lang] + 1
		}
	}

	return res, stats
}

type byKey struct {
	messages []*message
}

func (k *byKey) Len() int {
	return len(k.messages)
}

func (k *byKey) Swap(i, j int) {
	k.messages[i], k.messages[j] = k.messages[j], k.messages[i]
}

func (k *byKey) Less(i, j int) bool {
	return k.messages[i].Key < k.messages[j].Key
}

func write(filename string, new []*message) error {
	sort.Sort(&byKey{new})

	content, err := json.MarshalIndent(new, "", "  ")
	if err != nil {
		return err
	}

	content = append(content, []byte("\n\r")...)

	err = ioutil.WriteFile(filename, content, 0644)
	if err != nil {
		return err
	}

	return nil
}

func updateMessages(filename string) error {
	old, err := read(filename)
	if err != nil {
		return err
	}

	new := errors.GetAll()
	merged, stats := merge(old, new)

	err = write(filename, merged)
	if err != nil {
		return err
	}

	f := "%10s %12s %12s\n"
	fmt.Println()
	fmt.Printf(f, "lang", "messages", "missing")
	total := stats[defaultLanguage]
	for lang, msgs := range stats {
		fmt.Printf(f, lang, fmt.Sprintf("%v", msgs), fmt.Sprintf("%v", total-msgs))
	}
	fmt.Println()

	return nil
}

func main() {
	mgr := config.Initialize("messages", cfg{
		Filename: "./messages.json",
	})

	err := mgr.Parse(os.Args...)
	if err != nil {
		log.WithError(err).Fatal("Could not parse options")
	}

	cfg := new(cfg)
	err = mgr.Unmarshal(&cfg)
	if err != nil {
		log.WithError(err).Fatal("Could not parse options")
	}

	ErrSomeUserMistake := &errors.ErrDescriptor{
		MessageFormat: "You made a mistake cost us {price, plural, =0 {no money} =1 {one dollar} other {{price} dollars}}",
		Type:          errors.InvalidArgument,
		Code:          391,
	}
	ErrSomeUserMistake.Register()

	log.WithField("Filename", cfg.Filename).Info("Updating messages")

	err = updateMessages(cfg.Filename)
	if err != nil {
		log.WithError(err).Fatal("Could not update messages")
	}
}
