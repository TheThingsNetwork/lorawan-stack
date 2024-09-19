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

package templates

import (
	"context"
	"fmt"

	"go.thethings.network/lorawan-stack/v3/pkg/email"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func init() {
	tmpl, err := email.NewTemplateFS(
		fsys, ttnpb.NotificationType_API_KEY_CHANGED,
		email.FSTemplate{
			SubjectTemplate:      "An API key of your {{ .Notification.EntityIds.EntityType }} has been changed",
			HTMLTemplateBaseFile: "base.html.tmpl",
			HTMLTemplateFile:     "api_key_changed.html.tmpl",
			TextTemplateFile:     "api_key_changed.txt.tmpl",
		},
	)
	if err != nil {
		panic(err)
	}
	email.RegisterTemplate(tmpl)
	email.RegisterNotification(ttnpb.NotificationType_API_KEY_CHANGED, &email.NotificationBuilder{
		EmailTemplateName: ttnpb.NotificationType_API_KEY_CHANGED,
		DataBuilder:       newAPIKeyChangedData,
	})
}

func newAPIKeyChangedData(_ context.Context, data email.NotificationTemplateData) (email.NotificationTemplateData, error) {
	var nData ttnpb.APIKey
	if err := data.Notification().GetData().UnmarshalTo(&nData); err != nil {
		return nil, err
	}
	return &APIKeyChangedData{
		NotificationTemplateData: data,
		APIKey:                   &nData,
	}, nil
}

// APIKeyChangedData is the data for the api_key_changed email.
type APIKeyChangedData struct {
	email.NotificationTemplateData
	*ttnpb.APIKey
}

// ConsoleURL returns the URL to the API key in the Console.
func (a *APIKeyChangedData) ConsoleURL() string {
	if a.Notification().GetEntityIds().EntityType() == "user" {
		return fmt.Sprintf("%s/user/api-keys/%s", a.Network().ConsoleURL, a.APIKey.GetId())
	}
	return fmt.Sprintf("%s/api-keys/%s", a.NotificationTemplateData.ConsoleURL(), a.APIKey.GetId())
}
