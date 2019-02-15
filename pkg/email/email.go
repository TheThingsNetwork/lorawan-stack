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

// Package email provides an interface to send messages over email.
package email

// Sender is the interface for sending messages over email.
type Sender interface {
	Send(message *Message) error
}

// Config for sending emails
type Config struct {
	SenderName    string `name:"sender-name" description:"The name of the sender"`
	SenderAddress string `name:"sender-address" description:"The address of the sender"`
	Provider      string `name:"provider" description:"Email provider to use"`
	Network       struct {
		Name              string `name:"name" description:"The name of the network"`
		IdentityServerURL string `name:"identity-server-url" description:"The URL of the Identity Server"`
		ConsoleURL        string `name:"console-url" description:"The URL of the Console"`
	} `name:"network" description:"The network of the sender"`
}
