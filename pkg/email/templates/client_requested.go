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
		fsys, ttnpb.NotificationType_CLIENT_REQUESTED,
		email.FSTemplate{
			SubjectTemplate:      "A new OAuth client has been requested",
			HTMLTemplateBaseFile: "base.html.tmpl",
			HTMLTemplateFile:     "client_requested.html.tmpl",
			TextTemplateFile:     "client_requested.txt.tmpl",
		},
	)
	if err != nil {
		panic(err)
	}
	email.RegisterTemplate(tmpl)
	email.RegisterNotification(ttnpb.NotificationType_CLIENT_REQUESTED, &email.NotificationBuilder{
		EmailTemplateName: ttnpb.NotificationType_CLIENT_REQUESTED,
		DataBuilder:       newClientRequestedData,
	})
}

func newClientRequestedData(_ context.Context, data email.NotificationTemplateData) (email.NotificationTemplateData, error) {
	var emailMsg ttnpb.CreateClientEmailMessage
	if err := data.Notification().GetData().UnmarshalTo(&emailMsg); err != nil {
		return nil, err
	}
	return &ClientRequestedData{
		NotificationTemplateData: data,
		CreateClientRequest:      emailMsg.GetCreateClientRequest(),
		APIKey:                   emailMsg.GetApiKey(),
	}, nil
}

// ClientRequestedData is the data for the client_requested email.
type ClientRequestedData struct {
	email.NotificationTemplateData
	*ttnpb.CreateClientRequest
	*ttnpb.APIKey
}

func (crd *ClientRequestedData) getAPIKeyName() string {
	if crd.APIKey.GetName() == "" {
		return crd.APIKey.GetId()
	}
	return crd.APIKey.GetName()
}

// SenderType returns the type of the entity that triggered the email.
func (crd *ClientRequestedData) SenderType() string {
	if crd.Notification().GetSenderIds().IDString() == "" {
		return "API key"
	}
	return "User"
}

// SenderTypeMidSentence returns the type of the entity that triggered the email, altered to fit midsentence.
func (crd *ClientRequestedData) SenderTypeMidSentence() string {
	if crd.Notification().GetSenderIds().IDString() == "" {
		return "API key"
	}
	return "user"
}

// Sender returns the name of the User or APIKey used for triggering the email.
func (crd *ClientRequestedData) Sender() string {
	if crd.Notification().GetSenderIds().IDString() == "" {
		return crd.getAPIKeyName()
	}
	return crd.Notification().GetSenderIds().IDString()
}
