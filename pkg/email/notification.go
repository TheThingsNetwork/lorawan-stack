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

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type notificationRegistryCtxKeyType struct{}

var notificationRegistryCtxKey notificationRegistryCtxKeyType

func notificationRegistryFromContext(ctx context.Context) (NotificationRegistry, bool) {
	reg, ok := ctx.Value(notificationRegistryCtxKey).(NotificationRegistry)
	return reg, ok
}

func newContextWithNotificationRegistry(parent context.Context, reg NotificationRegistry) context.Context {
	return context.WithValue(parent, notificationRegistryCtxKey, reg)
}

// NotificationRegistry keeps track of email notifications.
type NotificationRegistry interface {
	RegisteredNotifications() []string
	GetNotification(ctx context.Context, name string) *NotificationBuilder
}

// NewNotificationRegistry returns a new empty email notification registry.
func NewNotificationRegistry() MapNotificationRegistry {
	return make(MapNotificationRegistry)
}

// MapNotificationRegistry is an email notification registry implementation.
type MapNotificationRegistry map[string]*NotificationBuilder

// RegisterNotification registers an email notification.
func (reg MapNotificationRegistry) RegisterNotification(name string, builder *NotificationBuilder) {
	reg[name] = builder
}

// RegisteredNotifications returns a sorted list of the names of all registered email notifications.
func (reg MapNotificationRegistry) RegisteredNotifications() []string {
	names := make([]string, 0, len(reg))
	for name := range reg {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// GetNotification returns a registered email notification from the registry.
func (reg MapNotificationRegistry) GetNotification(_ context.Context, name string) *NotificationBuilder {
	return reg[name]
}

var defaultNotificationRegistry = make(MapNotificationRegistry)

// RegisterNotification registers an email notification on the default registry.
func RegisterNotification(name string, builder *NotificationBuilder) {
	defaultNotificationRegistry.RegisterNotification(name, builder)
}

// RegisteredNotifications returns the names of the registered email notifications in the default registry.
func RegisteredNotifications() []string {
	return defaultNotificationRegistry.RegisteredNotifications()
}

// GetNotification returns a registered email notification from the registry in the context (if available), otherwise falling back to the default registry.
func GetNotification(ctx context.Context, name string) *NotificationBuilder {
	if reg, ok := notificationRegistryFromContext(ctx); ok {
		if tmpl := reg.GetNotification(ctx, name); tmpl != nil {
			return tmpl
		}
	}
	return defaultNotificationRegistry.GetNotification(ctx, name)
}

// NotificationTemplateData extends TemplateData for notifications.
type NotificationTemplateData interface {
	TemplateData
	Notification() *ttnpb.Notification
	ConsoleURL() string
	Receivers() string
}

// NotificationTemplateDataBuilder is used to extend NotificationTemplateData.
type NotificationTemplateDataBuilder func(context.Context, NotificationTemplateData) (NotificationTemplateData, error)

// NewNotificationTemplateData returns new notification template data.
func NewNotificationTemplateData(data TemplateData, notification *ttnpb.Notification) NotificationTemplateData {
	return &notificationTemplateData{
		TemplateData: data,
		notification: notification,
	}
}

type notificationTemplateData struct {
	TemplateData
	notification *ttnpb.Notification
}

func (d *notificationTemplateData) Notification() *ttnpb.Notification { return d.notification }

func (d *notificationTemplateData) ConsoleURL() string {
	url := strings.TrimSuffix(d.Network().ConsoleURL, "/")
	switch ids := d.notification.GetEntityIds().GetIds().(type) {
	case *ttnpb.EntityIdentifiers_ApplicationIds:
		url = fmt.Sprintf("%s/applications/%s", url, ids.ApplicationIds.GetApplicationId())
	case *ttnpb.EntityIdentifiers_DeviceIds:
		url = fmt.Sprintf("%s/applications/%s/devices/%s", url, ids.DeviceIds.GetApplicationIds().GetApplicationId(), ids.DeviceIds.GetDeviceId())
	case *ttnpb.EntityIdentifiers_GatewayIds:
		url = fmt.Sprintf("%s/gateways/%s", url, ids.GatewayIds.GetGatewayId())
	case *ttnpb.EntityIdentifiers_OrganizationIds:
		url = fmt.Sprintf("%s/organizations/%s", url, ids.OrganizationIds.GetOrganizationId())
	case *ttnpb.EntityIdentifiers_UserIds:
		if d.Receiver().GetAdmin() {
			url = fmt.Sprintf("%s/admin-panel/user-management/%s", url, ids.UserIds.GetUserId())
		}
	}
	return url
}

func (d *notificationTemplateData) Receivers() string {
	receivers := d.Notification().GetReceivers()
	if len(receivers) == 0 {
		return ""
	}
	receiverStrings := make([]string, 0, len(receivers))
	for _, receiver := range receivers {
		switch receiver {
		case ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_ADMINISTRATIVE_CONTACT:
			receiverStrings = append(receiverStrings, "administrative contacts")
		case ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_TECHNICAL_CONTACT:
			receiverStrings = append(receiverStrings, "technical contacts")
		case ttnpb.NotificationReceiver_NOTIFICATION_RECEIVER_COLLABORATOR:
			receiverStrings = append(receiverStrings, "collaborators")
		}
	}
	switch len(receiverStrings) {
	case 0:
		return ""
	case 1:
		return receiverStrings[0]
	default:
		return fmt.Sprintf(
			"%s and %s",
			strings.Join(receiverStrings[:len(receiverStrings)-1], ", "),
			receiverStrings[len(receiverStrings)-1],
		)
	}
}

// NotificationBuilder is used to build notifications.
type NotificationBuilder struct {
	EmailTemplateName string
	DataBuilder       NotificationTemplateDataBuilder
}
