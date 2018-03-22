// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb

import (
	"math"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
)

// Validate is used as validator function by the GRPC validator interceptor.
func (s TxSettings) Validate() error {
	if s.GetChannelIndex() > math.MaxUint8 {
		return ErrTxChIdxTooHigh.New(nil)
	}

	if s.GetDataRateIndex() > math.MaxUint8 {
		return ErrTxDRIdxTooHigh.New(nil)
	}
	return nil
}

// Validate is used as validator function by the GRPC validator interceptor.
func (p MACPayload) Validate() error {
	if p.DevAddr.IsZero() {
		return ErrMissingDevAddr.New(nil)
	}

	if p.GetFCnt() > math.MaxUint16 {
		return ErrFCntTooHigh.New(nil)
	}

	if p.FPort > math.MaxUint8 {
		return ErrFPortTooHigh.New(nil)
	}

	return nil
}

// Validate is used as validator function by the GRPC validator interceptor.
func (p JoinRequestPayload) Validate() error {
	if p.DevEUI.IsZero() {
		return ErrMissingDevEUI.New(nil)
	}
	if p.JoinEUI.IsZero() {
		return ErrMissingJoinEUI.New(nil)
	}

	return nil
}

// Validate is used as validator function by the GRPC validator interceptor.
func (p RejoinRequestPayload) Validate() error {
	// TODO: implement
	return nil
}

// Validate is used as validator function by the GRPC validator interceptor.
func (m UplinkMessage) Validate() error {
	if err := m.GetSettings().Validate(); err != nil {
		return err
	}

	if len(m.GetRawPayload()) == 0 {
		return ErrMissingRawPayload.New(nil)
	}

	p := m.GetPayload()
	switch p.GetMType() {
	case MType_CONFIRMED_UP, MType_UNCONFIRMED_UP:
		mp := p.GetMACPayload()
		if mp == nil {
			return ErrMissingPayload.New(nil)
		}
		return mp.Validate()
	case MType_JOIN_REQUEST:
		jp := p.GetJoinRequestPayload()
		if jp == nil {
			return ErrMissingPayload.New(nil)
		}
		return jp.Validate()
	case MType_REJOIN_REQUEST:
		rp := p.GetRejoinRequestPayload()
		if rp == nil {
			return ErrMissingPayload.New(nil)
		}
		return rp.Validate()
	default:
		return ErrWrongPayloadType.New(errors.Attributes{
			"type": p.GetMType(),
		})
	}
}
