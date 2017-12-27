// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package mac

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

// HandleLinkCheckReq handles the LinkCheckReq MAC command
func HandleLinkCheckReq(ctx context.Context, dev *ttnpb.EndDevice, cmd *ttnpb.MACCommand) error {
	rxMetadatas := rxMetadataFromContext(ctx)
	if len(rxMetadatas) == 0 {
		panic(errors.New("No RxMetadata in uplink message"))
	}

	txSettings := txSettingsFromContext(ctx)
	if txSettings == nil {
		panic(errors.New("No TxSettings in uplink message"))
	}

	floor, ok := demodulationFloor[sfbw{txSettings.SpreadingFactor, txSettings.Bandwidth}]
	if !ok {
		return errors.New("Invalid data rate")
	}

	bestSNR := rxMetadatas[0].GetSNR()
	gateways := make(map[string]struct{}, len(rxMetadatas))

	for _, meta := range rxMetadatas {
		gateways[meta.GatewayID] = struct{}{}
		if meta.SNR > bestSNR {
			bestSNR = meta.SNR
		}
	}

	ans := &ttnpb.MACCommand_LinkCheckAns{
		Margin:       uint32(uint8(bestSNR - floor)),
		GatewayCount: uint32(len(gateways)),
	}

	dequeueMAC(dev, ttnpb.CID_LINK_CHECK)
	enqueueMAC(dev, ans.MACCommand())

	return nil
}

func init() {
	RegisterHandler(ttnpb.CID_LINK_CHECK, UplinkHandlerFunc(HandleLinkCheckReq))
}
