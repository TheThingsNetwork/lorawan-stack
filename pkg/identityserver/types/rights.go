// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

// Right is the type that represents a right to do something on TTN
type Right string

const (
	// ApplicationSettingsRight is the Right to read and write access to the settings, devices and access keys of the application.
	ApplicationSettingsRight Right = "application:settings"

	// ApplicationCollaboratorsRight is the Right to edit and modify collaborators of the application.
	ApplicationCollaboratorsRight Right = "application:collaborators"

	// ApplicationDeleteRight is the Right to delete the application.
	ApplicationDeleteRight Right = "application:delete"

	// ReadUplinkRight is the Right to view messages sent by devices of the application.
	ReadUplinkRight Right = "messages:up:r"

	// WriteUplinkRight is the Right to send messages to the application.
	WriteUplinkRight Right = "messages:up:w"

	// WriteDownlinkRight is the Right to send messages to devices of the application.
	WriteDownlinkRight Right = "messages:down:w"

	// DevicesRight is the Right to list, edit and remove devices for the application on a handler.
	DevicesRight Right = "devices"

	// GatewayOwnerRight is the Right that states that a collaborator is an owner.
	GatewayOwnerRight Right = "gateway:owner"

	// GatewaySettingsRight is the Right to read and write access to the gateway settings.
	GatewaySettingsRight Right = "gateway:settings"

	// GatewayCollaboratorsRight is the Right to edit the gateway collaborators.
	GatewayCollaboratorsRight Right = "gateway:collaborators"

	// GatewayDeleteRight is the Right to delete a gateway.
	GatewayDeleteRight Right = "gateway:delete"

	// GatewayLocationRight is the Right to view the exact location of the gateway, otherwise only approximate location will be shown.
	GatewayLocationRight Right = "gateway:location"

	// GatewayStatusRight is the Right to view the gateway status and metrics about the gateway.
	GatewayStatusRight Right = "gateway:status"

	// GatewayMessagesRight is the Right to view the messages of a gateway.
	GatewayMessagesRight Right = "gateway:messages"

	// ComponentSettingsRight is the Right to read and write access to the settings and access key of a network component.
	ComponentSettingsRight Right = "component:settings"

	// ComponentDeleteRight is the Right to delete the network component.
	ComponentDeleteRight Right = "component:delete"

	// ComponentCollaboratorsRight is the Right to view and edit component collaborators.
	ComponentCollaboratorsRight Right = "component:collaborators"

	// ClientOwnerRight is the Right that states that a collaborator is an owner.
	ClientOwnerRight Right = "client:owner"

	// ClientSettingsRight is the Right to read and write the settings of a given client.
	ClientSettingsRight Right = "client:settings"

	// ClientDeleteRight is the Right to delete the client.
	ClientDeleteRight Right = "client:delete"

	// ClientCollaboratorsRight is the Right to view and edit the client collaborators.
	ClientCollaboratorsRight = "client:collaborators"
)

// String implements fmt.Stringer interface.
func (r Right) String() string {
	return string(r)
}
