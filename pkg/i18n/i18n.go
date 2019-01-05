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

// Package i18n helps with internationalization of registered messages.
package i18n

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const defaultLanguage = "en" // The language of the messages written in Go files.

// MessageDescriptor describes a translatable message.
type MessageDescriptor struct {
	Translations map[string]string `json:"translations,omitempty"`
	Description  struct {
		Package string `json:"package,omitempty"`
		File    string `json:"file,omitempty"`
	} `json:"description,omitempty"`
	touched bool
	updated bool
}

// Touched returns whether the descriptor was touched (i.e. it is still used).
func (m *MessageDescriptor) Touched() bool { return m.touched }

// Updated returns whether the descriptor was updated.
func (m *MessageDescriptor) Updated() bool { return m.updated }

// SetSource sets the source package and file name of the message descriptor.
// The argument skip is the number of stack frames to ascend, with 0 identifying the caller of SetSource.
func (m *MessageDescriptor) SetSource(skip uint) {
	_, file, _, ok := runtime.Caller(1 + int(skip))
	if !ok {
		panic("could not determine source of message")
	}
	m.Description.Package = filepath.Dir(file)
	if strings.Contains(m.Description.Package, "go.thethings.network/lorawan-stack/") {
		split := strings.Split(m.Description.Package, "go.thethings.network/lorawan-stack/")
		m.Description.Package = split[len(split)-1]
	}
	m.Description.File = filepath.Base(file)
}

// MessageDescriptorMap is a map of message descriptors.
type MessageDescriptorMap map[string]*MessageDescriptor

// Global registry.
var Global = make(MessageDescriptorMap)

// ReadFile reads the descriptors from a file.
func ReadFile(filename string) (MessageDescriptorMap, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	descriptors := make(MessageDescriptorMap)
	err = json.Unmarshal(bytes, &descriptors)
	if err != nil {
		return nil, err
	}
	return descriptors, nil
}

// MarshalJSON marshals the descriptors to JSON.
func (m MessageDescriptorMap) MarshalJSON() ([]byte, error) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b bytes.Buffer
	if err := b.WriteByte('{'); err != nil {
		return nil, err
	}
	for i, k := range keys {
		if i > 0 {
			if _, err := b.Write([]byte{','}); err != nil {
				return nil, err
			}
		}
		keyJSON, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		if _, err = b.Write(keyJSON); err != nil {
			return nil, err
		}
		if err = b.WriteByte(':'); err != nil {
			return nil, err
		}
		valJSON, err := json.Marshal(m[k])
		if err != nil {
			return nil, err
		}
		if _, err = b.Write(valJSON); err != nil {
			return nil, err
		}
	}
	if err := b.WriteByte('}'); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// WriteFile writes the descriptors to a file.
func (m MessageDescriptorMap) WriteFile(filename string) error {
	bytes, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, append(bytes, '\n'), 0644)
}

// Define a message.
func (m MessageDescriptorMap) Define(id, message string) *MessageDescriptor {
	if m[id] != nil {
		panic(fmt.Errorf("Message %s already defined", id))
	}
	m[id] = &MessageDescriptor{
		Translations: map[string]string{
			defaultLanguage: message,
		},
		touched: true,
	}
	m[id].SetSource(1)
	return m[id]
}

// Define a message in the global registry.
func Define(id, message string) *MessageDescriptor {
	d := Global.Define(id, message)
	d.SetSource(1)
	return d
}

// Merge messages from the given descriptor map into the current registry.
func (m MessageDescriptorMap) Merge(other MessageDescriptorMap) {
	for id, other := range other {
		if m[id] == nil {
			m[id] = other
		} else {
			if other.Translations[defaultLanguage] != m[id].Translations[defaultLanguage] {
				m[id].updated = true
			}
			for language, translation := range other.Translations {
				if language == defaultLanguage {
					continue // This one is set from the Define.
				}
				m[id].Translations[language] = translation
			}
		}
	}
}

// Updated returns updated message descriptors.
func (m MessageDescriptorMap) Updated() (updated []string) {
	for id, descriptor := range m {
		if descriptor.Updated() {
			updated = append(updated, id)
		}
	}
	return
}

// Cleanup removes unused message descriptors.
func (m MessageDescriptorMap) Cleanup() (deleted []string) {
	for id, descriptor := range m {
		if !descriptor.Touched() {
			delete(m, id)
			deleted = append(deleted, id)
		}
	}
	return
}
