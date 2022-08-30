// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mac

import (
	"sort"

	"go.thethings.network/lorawan-stack/v3/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type EnqueueState struct {
	MaxDownLen, MaxUpLen uint16
	QueuedEvents         events.Builders
	Ok                   bool
}

// enqueueMACCommand appends commands returned by f to cmds.
// Arguments to f represent the amount of downlink and uplink messages respectively with CID cid which fit in byte limits maxDownLen and maxUpLen.
// f returns a slice downlink commands to append to cmds, amount of uplinks to expect and bool indicating whether all commands fit.
// enqueueMACCommand returns the resulting downlink MAC command slice, new value for maxDownLen, maxUpLen and bool indicating whether all commands fit.
func enqueueMACCommand(cid ttnpb.MACCommandIdentifier, maxDownLen, maxUpLen uint16, f func(nDown, nUp uint16) ([]*ttnpb.MACCommand, uint16, events.Builders, bool), cmds ...*ttnpb.MACCommand) ([]*ttnpb.MACCommand, EnqueueState) {
	desc := lorawan.DefaultMACCommands[cid]
	maxDown := maxDownLen / (1 + desc.DownlinkLength)
	maxUp := maxUpLen / (1 + desc.UplinkLength)
	enq, nUp, evs, ok := f(maxDown, maxUp)
	if len(enq) > int(maxDown) || nUp > maxUp {
		panic("invalid amount of MAC commands enqueued")
	}
	return append(cmds, enq...), EnqueueState{
		MaxDownLen:   maxDownLen - uint16(len(enq))*(1+desc.DownlinkLength),
		MaxUpLen:     maxUpLen - nUp*(1+desc.UplinkLength),
		QueuedEvents: evs,
		Ok:           ok,
	}
}

// handleMACResponse searches for first command in cmds with CID equal to cid and calls f with found value as argument.
// handleMACResponse returns cmds with first MAC command with CID equal to cid removed or
// cmds passed and error if f returned non-nil error or if command with CID cid is not found in cmds
// and allowMissing is false.
func handleMACResponse(
	cid ttnpb.MACCommandIdentifier,
	allowMissing bool,
	f func(*ttnpb.MACCommand) error, cmds ...*ttnpb.MACCommand,
) ([]*ttnpb.MACCommand, error) {
	for i, cmd := range cmds {
		if cmd.Cid != cid {
			continue
		}
		if err := f(cmd); err != nil {
			return cmds, err
		}
		return append(cmds[:i], cmds[i+1:]...), nil
	}
	if allowMissing {
		return cmds, nil
	}
	return cmds, ErrRequestNotFound.WithAttributes("cid", cid)
}

// handleMACResponse searches for first MAC command block in cmds with CID equal to cid and calls f for each found value as argument.
// handleMACResponse returns cmds with first MAC command block with CID equal to cid removed or
// cmds passed and error if f returned non-nil error or if command with CID cid is not found in cmds
// and allowMissing is false.
func handleMACResponseBlock(
	cid ttnpb.MACCommandIdentifier,
	allowMissing bool,
	f func(*ttnpb.MACCommand) error,
	cmds ...*ttnpb.MACCommand,
) ([]*ttnpb.MACCommand, error) {
	first := -1
	last := -1

outer:
	for i, cmd := range cmds {
		last = i

		switch {
		case first >= 0 && cmd.Cid != cid:
			last--
			break outer
		case first < 0 && cmd.Cid != cid:
			continue
		case first < 0:
			first = i
		}

		if err := f(cmd); err != nil {
			return cmds, err
		}
	}
	switch {
	case first < 0 && allowMissing:
		return cmds, nil
	case first < 0 && !allowMissing:
		return cmds, ErrRequestNotFound.WithAttributes("cid", cid)
	default:
		return append(cmds[:first], cmds[last+1:]...), nil
	}
}

func searchDataRateIndex(v ttnpb.DataRateIndex, vs ...ttnpb.DataRateIndex) int {
	return sort.Search(len(vs), func(i int) bool { return vs[i] >= v })
}

func searchUint32(v uint32, vs ...uint32) int {
	return sort.Search(len(vs), func(i int) bool { return vs[i] >= v })
}

func searchUint64(v uint64, vs ...uint64) int {
	return sort.Search(len(vs), func(i int) bool { return vs[i] >= v })
}

func deviceRejectedADRDataRateIndex(dev *ttnpb.EndDevice, idx ttnpb.DataRateIndex) bool {
	i := searchDataRateIndex(idx, dev.MacState.RejectedAdrDataRateIndexes...)
	return i < len(dev.MacState.RejectedAdrDataRateIndexes) && dev.MacState.RejectedAdrDataRateIndexes[i] == idx
}

func deviceRejectedADRTXPowerIndex(dev *ttnpb.EndDevice, idx uint32) bool {
	i := searchUint32(idx, dev.MacState.RejectedAdrTxPowerIndexes...)
	return i < len(dev.MacState.RejectedAdrTxPowerIndexes) && dev.MacState.RejectedAdrTxPowerIndexes[i] == idx
}

func deviceRejectedFrequency(dev *ttnpb.EndDevice, freq uint64) bool {
	i := searchUint64(freq, dev.MacState.RejectedFrequencies...)
	return i < len(dev.MacState.RejectedFrequencies) && dev.MacState.RejectedFrequencies[i] == freq
}

func deviceRejectedDataRateRange(dev *ttnpb.EndDevice, freq uint64, min, max ttnpb.DataRateIndex) bool {
	for _, r := range dev.MacState.RejectedDataRateRanges[freq].GetRanges() {
		if r.MinDataRateIndex == min && r.MaxDataRateIndex == max {
			return true
		}
	}
	return false
}
