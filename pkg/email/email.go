// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

// Package email provides an interface to send messages over email.
package email

import "go.thethings.network/lorawan-stack/v3/pkg/ttnpb"

// Message for sending over email.
type Message struct {
	TemplateName ttnpb.NotificationType

	RecipientName    string
	RecipientAddress string

	Subject  string
	HTMLBody string
	TextBody string
}

// Sender is the interface for sending messages over email.
type Sender interface {
	Send(message *Message) error
}
