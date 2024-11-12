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
		fsys, ttnpb.GetNotificationTypeString(ttnpb.NotificationType_USER_REQUESTED),
		email.FSTemplate{
			SubjectTemplate:      "A new user has requested to join {{ .Network.Name }}",
			HTMLTemplateBaseFile: "base.html.tmpl",
			HTMLTemplateFile:     "user_requested.html.tmpl",
			TextTemplateFile:     "user_requested.txt.tmpl",
		},
	)
	if err != nil {
		panic(err)
	}
	email.RegisterTemplate(tmpl)
	email.RegisterNotification(
		ttnpb.GetNotificationTypeString(ttnpb.NotificationType_USER_REQUESTED), &email.NotificationBuilder{
			EmailTemplateName: ttnpb.GetNotificationTypeString(ttnpb.NotificationType_USER_REQUESTED),
			DataBuilder:       newUserRequestedData,
		})
}

func newUserRequestedData(_ context.Context, data email.NotificationTemplateData) (email.NotificationTemplateData, error) {
	var nData ttnpb.CreateUserRequest
	if err := data.Notification().GetData().UnmarshalTo(&nData); err != nil {
		return nil, err
	}
	return &UserRequestedData{
		NotificationTemplateData: data,
		CreateUserRequest:        &nData,
	}, nil
}

// UserRequestedData is the data for the user_requested email.
type UserRequestedData struct {
	email.NotificationTemplateData
	*ttnpb.CreateUserRequest
}
