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

	"go.thethings.network/lorawan-stack/v3/pkg/email"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func init() {
	tmpl, err := email.NewTemplateFS(
		fsys, ttnpb.NotificationType_PASSWORD_CHANGED,
		email.FSTemplate{
			SubjectTemplate:      "Your password on {{ .Network.Name }} has been changed",
			HTMLTemplateBaseFile: "base.html.tmpl",
			HTMLTemplateFile:     "password_changed.html.tmpl",
			TextTemplateFile:     "password_changed.txt.tmpl",
		},
	)
	if err != nil {
		panic(err)
	}
	email.RegisterTemplate(tmpl)
	email.RegisterNotification(ttnpb.NotificationType_PASSWORD_CHANGED, &email.NotificationBuilder{
		EmailTemplateName: ttnpb.NotificationType_PASSWORD_CHANGED,
		DataBuilder:       newPasswordChangedData,
	})
}

func newPasswordChangedData(_ context.Context, data email.NotificationTemplateData) (email.NotificationTemplateData, error) {
	return &PasswordChangedData{
		NotificationTemplateData: data,
	}, nil
}

// PasswordChangedData is the data for the password_changed email.
type PasswordChangedData struct {
	email.NotificationTemplateData
}
