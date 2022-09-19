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

package console

import (
	"go.thethings.network/lorawan-stack/v3/pkg/web/oauthclient"
	"go.thethings.network/lorawan-stack/v3/pkg/webui"
)

// UIConfig is the combined configuration for the Console UI.
type UIConfig struct {
	webui.TemplateData `name:",squash"`
	FrontendConfig     `name:",squash"`
}

// StackConfig is the configuration of the stack components.
type StackConfig struct {
	IS   webui.APIConfig `json:"is" name:"is"`
	GS   webui.APIConfig `json:"gs" name:"gs"`
	NS   webui.APIConfig `json:"ns" name:"ns"`
	AS   webui.APIConfig `json:"as" name:"as"`
	JS   webui.APIConfig `json:"js" name:"js"`
	EDTC webui.APIConfig `json:"edtc" name:"edtc"`
	QRG  webui.APIConfig `json:"qrg" name:"qrg"`
	GCS  webui.APIConfig `json:"gcs" name:"gcs"`
	DCS  webui.APIConfig `json:"dcs" name:"dcs"`
}

// FrontendConfig is the configuration for the Console frontend.
type FrontendConfig struct {
	DocumentationBaseURL   string `json:"documentation_base_url" name:"documentation-base-url" description:"The base URL for generating documentation links"`
	StatusPage             string `json:"status_page_base_url" name:"status-page-base-url" description:"The base URL for generating status page links"`
	Language               string `json:"language" name:"-"`
	SupportLink            string `json:"support_link" name:"support-link" description:"The URI that the support button will point to"`
	StackConfig            `json:"stack_config" name:",squash"`
	AccountURL             string `json:"account_url" name:"account-url" description:"The URL that points to the root of the Account"`
	DevEUIIssuingEnabled   bool   `json:"dev_eui_issuing_enabled" name:"dev-eui-issuing-enabled" description:"DevEUI issuer flag"`
	DevEUIApplicationLimit int    `json:"dev_eui_application_limit" name:"dev-eui-app-limit" description:"Limit on number of DevEUI's issued per application"`
}

// Config is the configuration for the Console.
type Config struct {
	OAuth oauthclient.Config `name:"oauth"`
	Mount string             `name:"mount" description:"Path on the server where the Console will be served"`
	UI    UIConfig           `name:"ui"`
}
