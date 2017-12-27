// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package mac

import (
	"context"

	"github.com/TheThingsNetwork/ttn/pkg/band"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

func joinLinkADRReq(reqs []*ttnpb.MACCommand_LinkADRReq, bandID band.ID) *ttnpb.MACCommand_LinkADRReq {
	var numChannels = 96 // TODO: Use lower numChannels depending on channel plan
	linkAdrReq := &ttnpb.MACCommand_LinkADRReq{
		ChannelMask: make([]bool, numChannels),
	}
	for _, req := range reqs {
		linkAdrReq.DataRateIndex = req.DataRateIndex
		linkAdrReq.TxPowerIndex = req.TxPowerIndex
		linkAdrReq.NbTrans = req.NbTrans
		switch req.ChannelMaskControl {
		case 6:
			switch bandID {
			case band.US_902_928, band.AU_915_928:
				// TODO: All 125kHz channels on
				// TODO: ChMask applies to channels 64 to 71
			default:
				// TODO: All defined channels on
			}
		case 7:
			switch bandID {
			case band.US_902_928, band.AU_915_928:
				// TODO: All 125kHz channels off
				// TODO: ChMask applies to channels 64 to 71
			default:
				// RFU
			}
		default:
			copy(linkAdrReq.ChannelMask[req.ChannelMaskControl*16:(req.ChannelMaskControl+1)*16], req.ChannelMask[:])
		}
	}
	return linkAdrReq
}

// LinkADRHandler handles the LinkAdr MAC command
type LinkADRHandler struct{}

// HandleMACCommand implements the Handler interface
func (*LinkADRHandler) HandleMACCommand(ctx context.Context, dev *ttnpb.EndDevice, cmd *ttnpb.MACCommand) error {
	linkAdrAns, ok := cmd.GetActualPayload().(*ttnpb.MACCommand_LinkADRAns)
	if !ok {
		return errors.Errorf("Expected *ttnpb.MACCommand_LinkADRAns payload, got %T", cmd.GetActualPayload())
	}

	var linkAdrReq *ttnpb.MACCommand_LinkADRReq
	requests := findMAC(dev, ttnpb.CID_LINK_ADR)
	switch len(requests) {
	case 0:
		return errors.New("Received LinkAdrAns without sending LinkAdrReq")
	case 1:
		linkAdrReq = requests[0].GetLinkAdrReq()
	default:
		switch dev.LoRaWANVersion {
		case ttnpb.MAC_V1_0, ttnpb.MAC_V1_0_1:
			// The answer applies to requests[0], there are len(requests) answers
			linkAdrReq = requests[0].GetLinkAdrReq()
			// TODO
		case ttnpb.MAC_V1_0_2:
			// The answer applies to requests[:], there are len(requests) answers
			linkAdrReq = requests[0].GetLinkAdrReq()
			// TODO
		case ttnpb.MAC_V1_1:
			// The answer applies to requests[:], there is one answer
			linkAdrReqs := make([]*ttnpb.MACCommand_LinkADRReq, len(requests))
			for i, req := range requests {
				linkAdrReqs[i] = req.GetLinkAdrReq()
			}
			linkAdrReq = joinLinkADRReq(linkAdrReqs, dev.FrequencyPlanID)
		}
	}

	if linkAdrAns.DataRateIndexAck && linkAdrAns.TxPowerIndexAck && linkAdrAns.ChannelMaskAck {
		dev.MACState.AdrDataRateIndex = linkAdrReq.DataRateIndex
		dev.MACState.AdrTxPowerIndex = linkAdrReq.TxPowerIndex
		// TODO: Set channel mask in MAC state
	}

	if !linkAdrAns.DataRateIndexAck {
		// TODO: Try something else?
	}
	if !linkAdrAns.TxPowerIndexAck {
		dev.MACState.AdrTxPowerIndex--
		if dev.MACState.AdrTxPowerIndex < 0 {
			dev.MACState.AdrTxPowerIndex = 0
			// TODO: Reconfigure channels maybe?
			// Could be that we're trying to set 125kHz data rate on 500kHz channels?
		}
	}
	if !linkAdrAns.ChannelMaskAck {
		// TODO: Reconfigure channels maybe?
	}

	dequeueMAC(dev, ttnpb.CID_LINK_ADR)

	return nil
}

// UpdateQueue implements the Handler interface
func (*LinkADRHandler) UpdateQueue(dev *ttnpb.EndDevice) error {
	// TODO: add the configured and enabled (not the same) channels to MACState
	// TODO: compare dev.MACState ADR params and enabled channels
	// TODO: build LinkADRReq(s), note the difference between 1.0, 1.0.1, 1.0.2 and 1.1 spec!
	return nil
}

func init() {
	RegisterHandler(ttnpb.CID_LINK_ADR, new(LinkADRHandler))
}
