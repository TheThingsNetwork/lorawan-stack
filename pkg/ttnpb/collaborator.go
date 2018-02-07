// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

// GetCollaboratorID returns the User's or Organization's ID that makes the
// application collaborationship.
func (c *ApplicationCollaborator) GetCollaboratorID() string {
	if userID := c.GetUserID(); userID != "" {
		return userID
	}
	return c.GetOrganizationID()
}

// GetCollaboratorID returns the User's or Organization's ID that makes the
// gateway collaborationship.
func (c *GatewayCollaborator) GetCollaboratorID() string {
	if userID := c.GetUserID(); userID != "" {
		return userID
	}
	return c.GetOrganizationID()
}
