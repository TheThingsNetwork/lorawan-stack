// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/gotnospirit/messageformat"
)

const defaultLanguage = "en" // The language of the messages written in Go files.

// MessageDescriptor describes a translatable message.
type MessageDescriptor struct {
	defaultFormat      *messageformat.MessageFormat
	Translations       map[string]string `json:"translations,omitempty"`
	translationFormats map[string]*messageformat.MessageFormat
	Description        struct {
		Package string `json:"package,omitempty"`
		File    string `json:"file,omitempty"`
	} `json:"description,omitempty"`
	id      string
	touched bool
	updated bool
}

// Load the messages
func (m *MessageDescriptor) Load() error {
	defaultParser, err := messageformat.NewWithCulture(defaultLanguage)
	if err != nil {
		return err
	}
	m.defaultFormat, err = defaultParser.Parse(m.Translations[defaultLanguage])
	if err != nil {
		return err
	}
	m.translationFormats = make(map[string]*messageformat.MessageFormat, len(m.Translations))
	for language, translation := range m.Translations {
		langParser, err := messageformat.NewWithCulture(language)
		if err != nil {
			return err
		}
		m.translationFormats[language], err = langParser.Parse(translation)
		if err != nil {
			return err
		}
	}
	return nil
}

// Format a message descriptor in the given language.
func (m *MessageDescriptor) Format(language string, data map[string]any) (msg string) {
	var err error
	if fmt := m.translationFormats[language]; fmt != nil {
		msg, err = fmt.FormatMap(data)
	} else {
		msg, err = m.defaultFormat.FormatMap(data)
	}
	if err != nil {
		msg = m.id // This shouldn't happen.
	}
	return
}

// Format a message from the global registry in the given language.
func Format(id, language string, data map[string]any) (msg string) {
	return Get(id).Format(language, data)
}

// Touched returns whether the descriptor was touched (i.e. it is still used).
func (m *MessageDescriptor) Touched() bool { return m.touched }

// Updated returns whether the descriptor was updated.
func (m *MessageDescriptor) Updated() bool { return m.updated }

func (m *MessageDescriptor) String() string {
	if m == nil {
		return "<nil message descriptor>"
	}
	if translation, ok := m.Translations[defaultLanguage]; ok {
		return translation
	}
	return m.id
}

var pathPrefix = func() string {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		panic("could not determine location of i18n.go")
	}
	return strings.TrimSuffix(file, filepath.Join("pkg", "i18n", "i18n.go"))
}()

// SetSource sets the source package and file name of the message descriptor.
// The argument skip is the number of stack frames to ascend, with 0 identifying the caller of SetSource.
func (m *MessageDescriptor) SetSource(skip uint) {
	_, file, _, ok := runtime.Caller(1 + int(skip))
	if !ok {
		panic("could not determine source of message")
	}
	m.Description.Package = strings.TrimPrefix(filepath.Dir(file), pathPrefix)
	m.Description.File = filepath.Base(file)
}

// MessageDescriptorMap is a map of message descriptors.
type MessageDescriptorMap map[string]*MessageDescriptor

var (
	global   = make(MessageDescriptorMap)
	globalMu = sync.RWMutex{}
)

// ReadFile reads the descriptors from a file.
func ReadFile(filename string) (MessageDescriptorMap, error) {
	bytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	descriptors := make(MessageDescriptorMap)
	err = json.Unmarshal(bytes, &descriptors)
	if err != nil {
		return nil, err
	}
	for id, descriptor := range descriptors {
		descriptor.id = id
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
	return os.WriteFile(filename, append(bytes, '\n'), 0o644) //nolint:gas
}

// Define a message.
func (m MessageDescriptorMap) Define(id, message string) *MessageDescriptor {
	if existing := m[id]; existing != nil {
		panic(fmt.Errorf("Message %s already defined in package %s (%s)", id, existing.Description.Package, existing.Description.File))
	}
	md := &MessageDescriptor{
		Translations: map[string]string{
			defaultLanguage: message,
		},
		id:      id,
		touched: true,
	}
	if err := md.Load(); err != nil {
		panic(err)
	}
	md.SetSource(1)
	m[id] = md
	return md
}

// Define a message in the global registry.
func Define(id, message string) *MessageDescriptor {
	globalMu.Lock()
	defer globalMu.Unlock()
	d := global.Define(id, message)
	d.SetSource(1)
	return d
}

// Get returns the MessageDescriptor of a specific message.
func (m MessageDescriptorMap) Get(id string) *MessageDescriptor {
	if md, ok := m[id]; ok {
		return md
	}
	return nil
}

// Get returns the MessageDescriptor of a specific message from the global registry.
func Get(id string) *MessageDescriptor {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return global.Get(id)
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

// CloneGlobal returns a shallow clone of the global message descriptor registry.
func CloneGlobal() MessageDescriptorMap {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return maps.Clone(global)
}
