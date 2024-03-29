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

package oauth

import (
	"go.thethings.network/lorawan-stack/v3/pkg/webui"
)

// UIConfig is the combined configuration for the OAuth and Account UI.
type UIConfig struct {
	webui.TemplateData `name:",squash"`
	FrontendConfig     `name:",squash"`
}

// StackConfig is the configuration of the stack components.
type StackConfig struct {
	IS webui.APIConfig `json:"is" name:"is"`
}

// FrontendConfig is the configuration for the OAuth frontend.
type FrontendConfig struct {
	DocumentationBaseURL   string `json:"documentation_base_url" name:"documentation-base-url" description:"The base URL for generating documentation links"`
	StatusPage             string `json:"status_page_base_url" name:"status-page-base-url" description:"The base URL for generating status page links"`
	Language               string `json:"language" name:"-"`
	StackConfig            `json:"stack_config" name:",squash"`
	EnableUserRegistration bool   `json:"enable_user_registration" name:"-"`
	ConsoleURL             string `json:"console_url" name:"console-url" description:"The URL that points to the root of the Console"`
}

// Config is the configuration for the OAuth server.
type Config struct {
	Mount       string   `name:"mount" description:"Path on the server where the Account application and OAuth services will be served"`
	UI          UIConfig `name:"ui"`
	CSRFAuthKey []byte   `name:"-"`
}
