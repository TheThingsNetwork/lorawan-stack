// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package mac

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// Handler of MAC commands
type Handler interface {
	// HandleMACCommand handles an uplink MAC command
	// A non-nil error may only be returned if that error must stop processing of the entire message.
	// Any non-critical errors should be logged to the context (TODO).
	HandleMACCommand(ctx context.Context, dev *ttnpb.EndDevice, payload *ttnpb.MACCommand) error
	// UpdateQueue updates the MAC command queue by comparing the active MAC state and the desired MAC state.
	// A non-nil error may only be returned if that error must stop processing of the entire message.
	UpdateQueue(dev *ttnpb.EndDevice) error
}

// UplinkHandlerFunc handles uplink MAC commands
type UplinkHandlerFunc func(ctx context.Context, dev *ttnpb.EndDevice, payload *ttnpb.MACCommand) error

// HandleMACCommand implements the Handler interface
func (f UplinkHandlerFunc) HandleMACCommand(ctx context.Context, dev *ttnpb.EndDevice, payload *ttnpb.MACCommand) error {
	return f(ctx, dev, payload)
}

// UpdateQueue implements the Handler interface. For UplinkHandlerFunc, it's a no-op.
func (UplinkHandlerFunc) UpdateQueue(dev *ttnpb.EndDevice) error { return nil }

var handlers = map[ttnpb.MACCommandIdentifier]Handler{}

// RegisterHandler registers a MAC command handler.
// The handler is called for each uplink message
func RegisterHandler(cid ttnpb.MACCommandIdentifier, handler Handler) {
	handlers[cid] = handler
}

// UpdateQueue updates the MAC command queue by comparing the active MAC state and the desired MAC state.
func UpdateQueue(dev *ttnpb.EndDevice) error {
	for _, handler := range handlers {
		if err := handler.UpdateQueue(dev); err != nil {
			return err
		}
	}
	return nil
}

func enqueueMAC(dev *ttnpb.EndDevice, cmd *ttnpb.MACCommand) {
	dev.QueuedMACCommands = append(dev.QueuedMACCommands, cmd)
}

func findMAC(dev *ttnpb.EndDevice, cid ttnpb.MACCommandIdentifier) (cmds []*ttnpb.MACCommand) {
	for _, existing := range dev.QueuedMACCommands {
		if existing.CID() == cid {
			cmds = append(cmds, existing)
		}
	}
	return
}

// dequeue all MAC commands with the given CID
func dequeueMAC(dev *ttnpb.EndDevice, cid ttnpb.MACCommandIdentifier) {
	updated := make([]*ttnpb.MACCommand, 0, len(dev.QueuedMACCommands))
	for _, existing := range dev.QueuedMACCommands {
		if existing.CID() != cid {
			updated = append(updated, existing)
		}
	}
	dev.QueuedMACCommands = updated
}

// dequeue the first occurence of a MAC command with the given CID
func dequeueFirstMAC(dev *ttnpb.EndDevice, cid ttnpb.MACCommandIdentifier) {
	updated := make([]*ttnpb.MACCommand, 0, len(dev.QueuedMACCommands))
	for i, existing := range dev.QueuedMACCommands {
		if existing.CID() == cid {
			updated = append(updated, dev.QueuedMACCommands[i+1:]...)
			break
		}
		updated = append(updated, existing)
	}
	dev.QueuedMACCommands = updated
}
