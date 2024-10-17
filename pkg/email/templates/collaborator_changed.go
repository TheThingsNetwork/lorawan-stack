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
		fsys, ttnpb.NotificationType_COLLABORATOR_CHANGED,
		email.FSTemplate{
			SubjectTemplate:      "A collaborator of your {{ .Notification.EntityIds.EntityType }} has been changed",
			HTMLTemplateBaseFile: "base.html.tmpl",
			HTMLTemplateFile:     "collaborator_changed.html.tmpl",
			TextTemplateFile:     "collaborator_changed.txt.tmpl",
		},
	)
	if err != nil {
		panic(err)
	}
	email.RegisterTemplate(tmpl)
	email.RegisterNotification(ttnpb.NotificationType_COLLABORATOR_CHANGED, &email.NotificationBuilder{
		EmailTemplateName: ttnpb.NotificationType_COLLABORATOR_CHANGED,
		DataBuilder:       newCollaboratorChangedData,
	})
}

func newCollaboratorChangedData(_ context.Context, data email.NotificationTemplateData) (email.NotificationTemplateData, error) {
	var nData ttnpb.Collaborator
	if err := data.Notification().GetData().UnmarshalTo(&nData); err != nil {
		return nil, err
	}
	return &CollaboratorChangedData{
		NotificationTemplateData: data,
		Collaborator:             &nData,
	}, nil
}

// CollaboratorChangedData is the data for the collaborator_changed email.
type CollaboratorChangedData struct {
	email.NotificationTemplateData
	*ttnpb.Collaborator
}

// ConsoleURL returns the URL to the API key in the Console.
func (d *CollaboratorChangedData) ConsoleURL() string {
	return fmt.Sprintf("%s/collaborators/%s/%s", d.NotificationTemplateData.ConsoleURL(), d.Collaborator.EntityType(), d.Collaborator.IDString())
}
