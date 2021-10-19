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

// Package i18n helps with internationalization of registered messages.
package i18n

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/gotnospirit/messageformat"
	jsoniter "github.com/json-iterator/go"
	"golang.org/x/text/language"
)

// Message is a translatable or translated message.
type Message struct {
	id      string
	message string
	format  *messageformat.MessageFormat

	pc        []uintptr
	arguments []string
}

// ID returns the ID of the message.
func (m *Message) ID() string { return m.id }

func (m *Message) String() string { return m.message }

// Format formats the message with the given data.
func (m *Message) Format(data map[string]interface{}) (string, error) {
	return m.format.FormatMap(data)
}

// Arguments returns the arguments in the message.
func (m *Message) Arguments() []string {
	if m.arguments == nil {
		m.arguments = messageFormatArguments(m.message)
	}
	return m.arguments
}

// Clone creates a copy of the message.
func (m *Message) Clone() *Message {
	return &Message{
		id:      m.id,
		message: m.message,
		format:  m.format,

		pc:        m.pc,
		arguments: m.arguments,
	}
}

// Messages is a collection of messages for a language.
type Messages struct {
	lang     language.Tag
	parser   *messageformat.Parser
	messages map[string]*Message
}

// NewMessages creates a new message collection for the given language.
func NewMessages(lang language.Tag) (*Messages, error) {
	parser, err := messageformat.NewWithCulture(lang.String())
	if err != nil {
		return nil, err
	}
	return &Messages{
		lang:     lang,
		parser:   parser,
		messages: make(map[string]*Message),
	}, nil
}

// Clone creates a copy of the messages.
func (m *Messages) Clone() *Messages {
	clonedMessages := make(map[string]*Message, len(m.messages))
	for id, message := range m.messages {
		clonedMessages[id] = message.Clone()
	}
	return &Messages{
		lang:     m.lang,
		parser:   m.parser,
		messages: clonedMessages,
	}
}

var i18nFile, stackPrefix = func() (string, string) {
	pc := make([]uintptr, 10)
	pc = pc[:runtime.Callers(1, pc)]
	frame, _ := runtime.CallersFrames(pc).Next()
	stackDir := filepath.Dir(filepath.Dir(filepath.Dir(frame.File)))
	return frame.File, stackDir + "/"
}()

func getCallers(pc []uintptr) string {
	frames := runtime.CallersFrames(pc)
	locs := make([]string, 0, len(pc))
	for {
		frame, more := frames.Next()
		if strings.HasPrefix(frame.File, stackPrefix) && frame.File != i18nFile {
			locs = append(locs, fmt.Sprintf("%s:%d", strings.TrimPrefix(frame.File, stackPrefix), frame.Line))
		}
		if !more {
			break
		}
	}
	return strings.Join(locs, " > ")
}

// Define defines a new message.
func (m *Messages) Define(id, message string) (*Message, error) {
	pc := make([]uintptr, 10)
	pc = pc[:runtime.Callers(1, pc)]
	if existing, exists := m.messages[id]; exists {
		return nil, fmt.Errorf("message %q defined at %s was previously defined at %s", id, getCallers(pc), getCallers(existing.pc))
	}
	format, err := m.parser.Parse(message)
	if err != nil {
		return nil, fmt.Errorf("message %q defined at %s could not be parsed: %w", id, getCallers(pc), err)
	}
	msg := &Message{
		id:      id,
		message: message,
		format:  format,

		pc: pc,
	}
	m.messages[id] = msg
	return msg, nil
}

// GetAllIDs returns the IDs of all messages.
func (m *Messages) GetAllIDs() []string {
	ids := make([]string, 0, len(m.messages))
	for id := range m.messages {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// Get returns a message by ID, or nil if not defined.
func (m *Messages) Get(id string) *Message {
	return m.messages[id]
}

// Delete deletes a message by ID if it exists.
func (m *Messages) Delete(id string) {
	delete(m.messages, id)
}

var codec = jsoniter.ConfigCompatibleWithStandardLibrary

// MarshalJSON implements the json.Marshaler interface.
func (m *Messages) MarshalJSON() ([]byte, error) {
	dto := make(map[string]string, len(m.messages))
	for id, msg := range m.messages {
		dto[id] = msg.message
	}
	return codec.Marshal(dto) // encoding/json sorts map keys.
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (m *Messages) UnmarshalJSON(data []byte) error {
	dto := make(map[string]string)
	if err := codec.Unmarshal(data, &dto); err != nil {
		return err
	}
	if m.messages == nil {
		m.messages = make(map[string]*Message)
	}
	for id, message := range dto {
		format, err := m.parser.Parse(message)
		if err != nil {
			return fmt.Errorf("message %q could not be parsed: %w", id, err)
		}
		msg, exists := m.messages[id]
		if !exists {
			msg = &Message{id: id}
			m.messages[id] = msg
		}
		msg.message = message
		msg.format = format
		msg.arguments = nil
	}
	return nil
}

// Format formats the message with the given ID with the given data.
// It returns an error if the message is not defined.
func (m *Messages) Format(id string, data map[string]interface{}) (string, error) {
	msg, ok := m.messages[id]
	if !ok {
		return "", fmt.Errorf("message %q not defined for language %q", id, m.lang)
	}
	return msg.Format(data)
}

// Bundle is a bundle of messages for different languages.
type Bundle struct {
	defaultLanguage language.Tag
	messages        map[language.Tag]*Messages
	matcher         language.Matcher
}

// NewBundle creates a new bundle of messages.
// Messages will be defined for the given default language.
func NewBundle(defaultLanguage language.Tag) (*Bundle, error) {
	defaultLanguageMessages, err := NewMessages(defaultLanguage)
	if err != nil {
		return nil, err
	}
	return &Bundle{
		defaultLanguage: defaultLanguage,
		messages: map[language.Tag]*Messages{
			defaultLanguage: defaultLanguageMessages,
		},
		matcher: language.NewMatcher([]language.Tag{defaultLanguage}),
	}, nil
}

// DefaultLanguage returns the default language of the bundle.
func (b *Bundle) DefaultLanguage() language.Tag { return b.defaultLanguage }

// Clone returns a copy of the bundle.
func (b *Bundle) Clone() *Bundle {
	clonedLanguages := make(map[language.Tag]*Messages, len(b.messages))
	for lang, messages := range b.messages {
		clonedLanguages[lang] = messages.Clone()
	}
	return &Bundle{
		defaultLanguage: b.defaultLanguage,
		messages:        clonedLanguages,
	}
}

// MessagesFor returns the message collection for the given language.
// If the bundle does not already contain a collection for this language, then
// it creates one, optionally filling it with fallback messages.
func (b *Bundle) MessagesFor(lang language.Tag, fallback bool) (*Messages, error) {
	messages, ok := b.messages[lang]
	if ok {
		return messages, nil
	}
	if fallback {
		fallbackLang, _, _ := b.matcher.Match(lang)
		messages = b.messages[fallbackLang].Clone()
	} else {
		var err error
		messages, err = NewMessages(lang)
		if err != nil {
			return nil, err
		}
	}
	b.messages[lang] = messages
	tags := make([]language.Tag, 0, len(b.messages))
	tags = append(tags, b.defaultLanguage)
	for lang := range b.messages {
		if lang != b.defaultLanguage {
			tags = append(tags, lang)
		}
	}
	b.matcher = language.NewMatcher(tags)
	return messages, nil
}

// Define defines a new message on the default language.
func (b *Bundle) Define(id, message string) (*Message, error) {
	return b.messages[b.defaultLanguage].Define(id, message)
}

// GetAllIDs returns the IDs of all messages defined for the default language.
func (b *Bundle) GetAllIDs() []string {
	return b.messages[b.defaultLanguage].GetAllIDs()
}

// Get returns a message by ID, or nil if not defined.
func (b *Bundle) Get(id string, lang language.Tag) *Message {
	messages, ok := b.messages[lang]
	if !ok {
		return nil
	}
	return messages.Get(id)
}

// MatchLanguage returns the first language that matches.
func (b *Bundle) MatchLanguage(preferences ...language.Tag) language.Tag {
	lang, _, _ := b.matcher.Match(preferences...)
	return lang
}

// MatchLanguageStrings matches the language by strings as commonly used in Accept-Language headers.
func (b *Bundle) MatchLanguageStrings(str ...string) language.Tag {
	lang, _ := language.MatchStrings(b.matcher, str...)
	return lang
}

// LoadFiles loads the locale files from the file system into the bundle.
func (b *Bundle) LoadFiles(fsys fs.FS) error {
	languageFiles, err := fs.Glob(fsys, "*.json")
	if err != nil {
		return err
	}
	for _, languageFile := range languageFiles {
		lang, err := language.Parse(filepath.Base(strings.TrimSuffix(languageFile, ".json")))
		if err != nil {
			return err
		}
		translations, err := b.MessagesFor(lang, true)
		if err != nil {
			return err
		}
		languageData, err := fs.ReadFile(fsys, languageFile)
		if err != nil {
			return err
		}
		err = json.Unmarshal(languageData, translations)
		if err != nil {
			return err
		}
	}
	return nil
}
