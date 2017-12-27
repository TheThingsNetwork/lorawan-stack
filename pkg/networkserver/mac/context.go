// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package mac

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

type contextKey string

var (
	uplinkKey     = contextKey("uplinkMessage")
	downlinkKey   = contextKey("downlinkMessage")
	txSettingsKey = contextKey("txSettings")
)

func newContextWithUplinkMessage(ctx context.Context, uplink *ttnpb.UplinkMessage) context.Context {
	ctx = context.WithValue(ctx, uplinkKey, uplink)
	ctx = context.WithValue(ctx, txSettingsKey, &uplink.Settings)
	return ctx
}

func newContextWithDownlinkMessage(ctx context.Context, downlink *ttnpb.DownlinkMessage) context.Context {
	ctx = context.WithValue(ctx, downlinkKey, downlink)
	ctx = context.WithValue(ctx, txSettingsKey, &downlink.Settings)
	return ctx
}

func rxMetadataFromContext(ctx context.Context) []ttnpb.RxMetadata {
	if uplink := ctx.Value(uplinkKey); uplink != nil {
		if uplink, ok := uplink.(*ttnpb.UplinkMessage); ok {
			return uplink.RxMetadata
		}
	}
	return nil
}

func txMetadataFromContext(ctx context.Context) *ttnpb.TxMetadata {
	if downlink := ctx.Value(downlinkKey); downlink != nil {
		if downlink, ok := downlink.(*ttnpb.DownlinkMessage); ok {
			return &downlink.TxMetadata
		}
	}
	return nil
}

func txSettingsFromContext(ctx context.Context) *ttnpb.TxSettings {
	if settings := ctx.Value(txSettingsKey); settings != nil {
		if settings, ok := settings.(*ttnpb.TxSettings); ok {
			return settings
		}
	}
	return nil
}
