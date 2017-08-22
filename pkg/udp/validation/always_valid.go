// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package validation

import "github.com/TheThingsNetwork/ttn/pkg/udp"

type alwaysValid struct{}

func (a alwaysValid) Valid(packet udp.Packet) bool { return true }

// AlwaysValid returns a Validator that considers all packets valid.
func AlwaysValid() Validator { return alwaysValid{} }
