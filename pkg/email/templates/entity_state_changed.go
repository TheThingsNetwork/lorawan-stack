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
		fsys, ttnpb.NotificationType_ENTITY_STATE_CHANGED,
		email.FSTemplate{
			SubjectTemplate:      "The state of your {{ .Notification.EntityIds.EntityType }} has been changed",
			HTMLTemplateBaseFile: "base.html.tmpl",
			HTMLTemplateFile:     "entity_state_changed.html.tmpl",
			TextTemplateFile:     "entity_state_changed.txt.tmpl",
		},
	)
	if err != nil {
		panic(err)
	}
	email.RegisterTemplate(tmpl)
	email.RegisterNotification(ttnpb.NotificationType_ENTITY_STATE_CHANGED, &email.NotificationBuilder{
		EmailTemplateName: ttnpb.NotificationType_ENTITY_STATE_CHANGED,
		DataBuilder:       newEntityStateChangedData,
	})
}

func newEntityStateChangedData(_ context.Context, data email.NotificationTemplateData) (email.NotificationTemplateData, error) {
	var nData ttnpb.EntityStateChangedNotification
	if err := data.Notification().GetData().UnmarshalTo(&nData); err != nil {
		return nil, err
	}
	return &EntityStateChangedData{
		NotificationTemplateData:       data,
		EntityStateChangedNotification: &nData,
	}, nil
}

// EntityStateChangedData is the data for the entity_state_changed email.
type EntityStateChangedData struct {
	email.NotificationTemplateData
	*ttnpb.EntityStateChangedNotification
}
