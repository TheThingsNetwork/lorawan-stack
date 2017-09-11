// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package utils

import "github.com/TheThingsNetwork/ttn/pkg/identityserver/types"

// Collaborator is a helper to construct a collaborator type.
func Collaborator(username string, rights []types.Right) types.Collaborator {
	return types.Collaborator{
		Username: username,
		Rights:   rights,
	}
}
