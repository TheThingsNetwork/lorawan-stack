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

package email

// NetworkConfig is the configuration of the network that sends the emails.
// This configuration is used by email templates.
type NetworkConfig struct {
	Name              string `name:"name" description:"The name of the network"`
	IdentityServerURL string `name:"identity-server-url" description:"The URL of the Identity Server"`
	ConsoleURL        string `name:"console-url" description:"The URL of the Console"`
	AssetsBaseURL     string `name:"assets-base-url" description:"The base URL to the email assets"`
	BrandingBaseURL   string `name:"branding-base-url" description:"The base URL to the email branding assets"`
}

// SenderConfig is the configuration of the sender.
type SenderConfig struct {
	SenderName    string `name:"sender-name" description:"The name of the sender"`
	SenderAddress string `name:"sender-address" description:"The address of the sender"`
}

// Config is the configuration for sending emails.
type Config struct {
	SenderConfig `name:",squash"`
	Network      NetworkConfig `name:"network" description:"The network of the sender"`
}
