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

// Package mock provides a test email provider that is used in tests.
package mock

import "go.thethings.network/lorawan-stack/pkg/email"

// Mock implements the email.Sender interface and stores sent emails internally.
type Mock struct {
	Messages []*email.Message
	Error    error
}

// New returns a new mock email.Sender.
func New() *Mock {
	return &Mock{}
}

// Send implements email.Sender.
// It appends the messages to Messages field and returns the Error field.
func (m *Mock) Send(message *email.Message) error {
	m.Messages = append(m.Messages, message)
	return m.Error
}
