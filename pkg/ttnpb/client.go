// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

// GetClient returns the base Client itself.
func (c *Client) GetClient() *Client {
	return c
}

// GetId implements osin.Client.
func (c *Client) GetId() string {
	return c.ClientIdentifier.GetClientID()
}

// GetRedirectUri implements osin.Client.
func (c *Client) GetRedirectUri() string {
	return c.RedirectURI
}

// GetUserData implements osin.Client.
func (c *Client) GetUserData() interface{} {
	return nil
}

func (c *Client) HasGrant(grant GrantType) bool {
	for _, g := range c.Grants {
		if g == grant {
			return true
		}
	}

	return false
}
