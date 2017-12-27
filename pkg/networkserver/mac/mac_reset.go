// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package mac

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/band"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/pkg/errors"
)

// TODO: Maybe this should be an exported func, as MAC state should also be reset on JoinAccept.
func resetMACState(state *ttnpb.MACState, band band.Band) {
	state.AdrAckDelay = uint32(band.AdrAckDelay)
	state.MaxTxPower = uint32(band.DefaultMaxEIRP)
	// state.UplinkDwellTime = // TODO
	// state.DownlinkDwellTime = // TODO
	state.AdrDataRateIndex = 0
	state.AdrTxPowerIndex = 0
	state.AdrNbTrans = 1
	state.AdrAckLimit = uint32(band.AdrAckLimit)
	state.AdrAckDelay = uint32(band.AdrAckDelay)
	state.DutyCycle = 0
	state.RxDelay = uint32(band.ReceiveDelay1.Seconds())
	state.Rx1DataRateOffset = 0
	state.Rx2DataRateIndex = uint32(band.DefaultRx2Parameters.DataRateIndex)
	state.Rx2Frequency = uint64(band.DefaultRx2Parameters.Frequency)
	state.RejoinTimer = 0
	state.RejoinCounter = 0
	state.PingSlotFrequency = 0
	state.PingSlotDataRateIndex = 0
}

// HandleResetInd handles the ResetInd MAC command
func HandleResetInd(ctx context.Context, dev *ttnpb.EndDevice, cmd *ttnpb.MACCommand) error {
	resetInd, ok := cmd.GetActualPayload().(*ttnpb.MACCommand_ResetInd)
	if !ok {
		return errors.Errorf("Expected *ttnpb.MACCommand_ResetInd payload, got %T", cmd.GetActualPayload())
	}
	switch resetInd.MinorVersion {
	case 1:
		dev.LoRaWANVersion = ttnpb.MAC_V1_1 // TODO: only if lower than 1.1
		band, err := band.GetByID(dev.FrequencyPlanID)
		if err != nil {
			return err
		}
		resetMACState(dev.MACState, band)
		conf := &ttnpb.MACCommand_ResetConf{MinorVersion: 1}
		dequeueMAC(dev, ttnpb.CID_RESET)
		enqueueMAC(dev, conf.MACCommand())
	}
	return nil
}

func init() {
	RegisterHandler(ttnpb.CID_RESET, UplinkHandlerFunc(HandleResetInd))
}
