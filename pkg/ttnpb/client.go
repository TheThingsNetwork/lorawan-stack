// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"crypto/subtle"
	"regexp"
)

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

// ClientSecretMatches implements osin.ClientSecretMatcher.
func (c *Client) ClientSecretMatches(secret string) bool {
	return subtle.ConstantTimeCompare([]byte(c.Secret), []byte(secret)) == 1
}

func (c *Client) HasGrant(grant GrantType) bool {
	for _, g := range c.Grants {
		if g == grant {
			return true
		}
	}

	return false
}

var (
	// FieldPathClientDescription is the field path for the client description field.
	FieldPathClientDescription = regexp.MustCompile(`^description$`)

	// FieldPathClientRedirectURI is the field path for the client redirect URI field.
	FieldPathClientRedirectURI = regexp.MustCompile(`^redirect_uri$`)

	// FieldPathClientRights is the field path for the client rights field.
	FieldPathClientRights = regexp.MustCompile(`^rights$`)
)
