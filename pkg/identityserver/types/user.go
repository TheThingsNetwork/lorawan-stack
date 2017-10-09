// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import "github.com/TheThingsNetwork/ttn/pkg/ttnpb"

// User is the interface of all things that can be an User.
// This can be used to build richer user types that can still be
// read and written to a database.
type User interface {
	// GetUser returns the ttnpb.User that represents this user.
	GetUser() *ttnpb.User
}
