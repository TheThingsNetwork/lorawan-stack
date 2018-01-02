// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import "github.com/TheThingsNetwork/ttn/pkg/ttnpb"

// Application is the interface of all things that can be an application.
// This can be used to build richer user types that can still be
// read and written to a database.
type Application interface {
	// GetApplication returns the ttnpb.Application that represents this
	// application.
	GetApplication() *ttnpb.Application
}
