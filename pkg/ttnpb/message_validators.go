// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package ttnpb

import (
	"context"
	"fmt"
	"math"

	"go.thethings.network/lorawan-stack/pkg/errors"
)

// ValidateContext is used as validator function by the GRPC validator interceptor.
// TODO: Adapt to genetrated validators (https://github.com/TheThingsIndustries/lorawan-stack/issues/1335)
func (s TxSettings) ValidateContext(ctx context.Context) error {
	if s.GetDataRateIndex() > math.MaxUint8 {
		return errExpectedLowerOrEqual("TxDRIdx", math.MaxUint8)(s.GetDataRateIndex())
	}
	return s.Validate()
}

// ValidateContext is used as validator function by the GRPC validator interceptor.
func (p MACPayload) ValidateContext(context.Context) error {
	if p.DevAddr.IsZero() {
		return errMissing("DevAddr")
	}
	return p.Validate()
}

// ValidateContext is used as validator function by the GRPC validator interceptor.
func (p JoinRequestPayload) ValidateContext(context.Context) error {
	if p.DevEUI.IsZero() {
		return errMissing("DevEUI")
	}
	if p.JoinEUI.IsZero() {
		return errMissing("JoinEUI")
	}
	return p.Validate()
}

var (
	errExpectedUplinkMType = errors.DefineInvalidArgument("expected_uplink_mtype", "MType `{result}` is not an uplink MType")
	errMissingRawPayload   = errors.DefineInvalidArgument("raw_payload", "missing raw payload")
)

// ValidateContext is used as validator function by the GRPC validator interceptor.
func (m *UplinkMessage) ValidateContext(context.Context) error {
	if p := m.GetPayload(); p.Payload == nil {
		if len(m.GetRawPayload()) == 0 {
			return errMissingRawPayload
		}
	} else {
		switch p.GetMType() {
		case MType_CONFIRMED_UP, MType_UNCONFIRMED_UP:
			mp := p.GetMACPayload()
			if mp == nil {
				return errMissing("MACPayload")
			}
			return mp.Validate()
		case MType_JOIN_REQUEST:
			jp := p.GetJoinRequestPayload()
			if jp == nil {
				return errMissing("JoinRequestPayload")
			}
			return jp.Validate()
		case MType_REJOIN_REQUEST:
			rp := p.GetRejoinRequestPayload()
			if rp == nil {
				return errMissing("RejoinRequestPayload")
			}
			return rp.Validate()
		default:
			return errExpectedUplinkMType.WithAttributes("result", p.GetMType().String())
		}
	}
	return m.Validate()
}

// ValidateContext reports whether cmd represents a valid *MACCommand.
func (cmd *MACCommand) ValidateContext(context.Context) error {
	return cmd.CID.Validate()
}

// Validate reports whether cid represents a valid MACCommandIdentifier.
// TODO: Move to generated validators (https://github.com/TheThingsIndustries/lorawan-stack/issues/1335)
func (cid MACCommandIdentifier) Validate() error {
	if cid < 0x00 || cid > 0xff {
		return errExpectedBetween("CID", "0x00", "0xFF")(fmt.Sprintf("0x%X", int32(cid)))
	}
	return nil
}
