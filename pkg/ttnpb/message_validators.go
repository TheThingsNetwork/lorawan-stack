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
	"math"

	"go.thethings.network/lorawan-stack/pkg/errors"
)

// Validate is used as validator function by the GRPC validator interceptor.
func (s TxSettings) Validate() error {
	if s.GetChannelIndex() > math.MaxUint8 {
		return errExpectedLowerOrEqual("TxChIdx", math.MaxUint8)(s.GetChannelIndex())
	}

	if s.GetDataRateIndex() > math.MaxUint8 {
		return errExpectedLowerOrEqual("TxDRIdx", math.MaxUint8)(s.GetDataRateIndex())
	}
	return nil
}

// Validate is used as validator function by the GRPC validator interceptor.
func (p MACPayload) Validate() error {
	if p.DevAddr.IsZero() {
		return errMissing("DevAddr")
	}

	if p.GetFCnt() > math.MaxUint16 {
		return errExpectedLowerOrEqual("FCnt", math.MaxUint16)(p.GetFCnt())
	}

	if p.FPort > math.MaxUint8 {
		return errExpectedLowerOrEqual("FPort", math.MaxUint8)(p.FPort)
	}

	return nil
}

// Validate is used as validator function by the GRPC validator interceptor.
func (p JoinRequestPayload) Validate() error {
	if p.DevEUI.IsZero() {
		return errMissing("DevEUI")
	}
	if p.JoinEUI.IsZero() {
		return errMissing("JoinEUI")
	}

	return nil
}

// Validate is used as validator function by the GRPC validator interceptor.
func (p RejoinRequestPayload) Validate() error {
	// TODO: implement
	return nil
}

var (
	errExpectedUplinkMType = errors.DefineInvalidArgument("expected_uplink_mtype", "MType `{result}` is not an uplink MType")
	errMissingRawPayload   = errors.DefineInvalidArgument("raw_payload", "missing raw payload")
)

// Validate is used as validator function by the GRPC validator interceptor.
func (m UplinkMessage) Validate() error {
	if err := m.GetSettings().Validate(); err != nil {
		return err
	}

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

	return nil
}
