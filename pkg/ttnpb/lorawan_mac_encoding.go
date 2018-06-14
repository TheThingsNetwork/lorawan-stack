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
	"bytes"
	"io"
	"io/ioutil"
	"math"
	"time"

	"go.thethings.network/lorawan-stack/pkg/encoding/lorawan"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/gpstime"
)

func newLengthUnequalError(want, got int) error {
	return errors.Errorf("expected length to equal %d, got %d", want, got)
}

func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

// AppendLoRaWAN appends the marshaled ResetInd payload to the slice.
func (cmd *MACCommand_ResetInd) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if cmd.MinorVersion > 15 {
		return nil, errors.Errorf("expected MinorVersion to be less or equal to 15, got %d", cmd.MinorVersion)
	}
	dst = append(dst, byte(cmd.MinorVersion))
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the ResetInd payload.
func (cmd *MACCommand_ResetInd) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 1 {
		return newLengthUnequalError(1, len(b))
	}
	cmd.MinorVersion = uint32(b[0] & 0xf)
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (pld *MACCommand_ResetInd_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return pld.ResetInd.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (pld *MACCommand_ResetInd_) UnmarshalLoRaWAN(b []byte) error {
	if pld.ResetInd == nil {
		pld.ResetInd = &MACCommand_ResetInd{}
	}
	return pld.ResetInd.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled ResetConf payload to the slice.
func (cmd *MACCommand_ResetConf) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if cmd.MinorVersion > 15 {
		return nil, errors.Errorf("expected MinorVersion to be less or equal to 15, got %d", cmd.MinorVersion)
	}
	dst = append(dst, byte(cmd.MinorVersion))
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the ResetConf payload.
func (cmd *MACCommand_ResetConf) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 1 {
		return newLengthUnequalError(1, len(b))
	}
	cmd.MinorVersion = uint32(b[0] & 0xf)
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_ResetConf_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.ResetConf.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_ResetConf_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.ResetConf.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled LinkCheckAns payload to the slice.
func (cmd *MACCommand_LinkCheckAns) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if cmd.Margin > 254 {
		return nil, errors.Errorf("expected Margin to be less or equal to 254, got %d", cmd.Margin)
	}
	if cmd.GatewayCount > 255 {
		return nil, errors.Errorf("expected GatewayCount to be less or equal to 255, got %d", cmd.GatewayCount)
	}
	dst = append(dst, byte(cmd.Margin), byte(cmd.GatewayCount))
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the LinkCheckAns payload.
func (cmd *MACCommand_LinkCheckAns) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 2 {
		return newLengthUnequalError(2, len(b))
	}
	cmd.Margin = uint32(b[0])
	cmd.GatewayCount = uint32(b[1])
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_LinkCheckAns_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.LinkCheckAns.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_LinkCheckAns_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.LinkCheckAns.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled LinkADRReq payload to the slice.
func (cmd *MACCommand_LinkADRReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if cmd.DataRateIndex > 15 {
		return nil, errors.Errorf("expected DataRateIndex to be less or equal to 15, got %d", cmd.DataRateIndex)
	}
	if cmd.TxPowerIndex > 15 {
		return nil, errors.Errorf("expected TxPowerIndex to be less or equal to 15, got %d", cmd.TxPowerIndex)
	}
	if len(cmd.ChannelMask) > 16 {
		return nil, errors.Errorf("expected ChannelMask to be shorter or equal to 16, got %d", len(cmd.ChannelMask))
	}
	if cmd.ChannelMaskControl > 7 {
		return nil, errors.Errorf("expected ChannelMaskControl to be less or equal to 7, got %d", cmd.ChannelMaskControl)
	}
	if cmd.NbTrans > 15 {
		return nil, errors.Errorf("expected NbTrans to be less or equal to 15, got %d", cmd.NbTrans)
	}
	dst = append(dst, byte((cmd.DataRateIndex&0xf)<<4)^byte(cmd.TxPowerIndex&0xf))
	chMask := make([]byte, 2)
	for i := uint8(0); i < 16 && i < uint8(len(cmd.ChannelMask)); i++ {
		chMask[i/8] = chMask[i/8] ^ boolToByte(cmd.ChannelMask[i])<<(i%8)
	}
	dst = append(dst, chMask...)
	dst = append(dst, byte((cmd.ChannelMaskControl&0x7)<<4)^byte(cmd.NbTrans&0xf))
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the LinkADRReq payload.
func (cmd *MACCommand_LinkADRReq) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 4 {
		return newLengthUnequalError(4, len(b))
	}
	cmd.DataRateIndex = uint32(b[0] >> 4)
	cmd.TxPowerIndex = uint32(b[0] & 0xf)
	var chMask [16]bool
	for i := uint8(0); i < 16; i++ {
		if (b[1+i/8]>>(i%8))&1 == 1 {
			chMask[i] = true
		}
	}
	cmd.ChannelMask = chMask[:]
	cmd.ChannelMaskControl = uint32((b[3] >> 4) & 0x7)
	cmd.NbTrans = uint32(b[3] & 0xf)
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_LinkADRReq_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.LinkADRReq.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_LinkADRReq_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.LinkADRReq.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled LinkADRAns payload to the slice.
func (cmd *MACCommand_LinkADRAns) AppendLoRaWAN(dst []byte) ([]byte, error) {
	var status byte
	if cmd.ChannelMaskAck {
		status |= 1
	}
	if cmd.DataRateIndexAck {
		status |= (1 << 1)
	}
	if cmd.TxPowerIndexAck {
		status |= (1 << 2)
	}
	dst = append(dst, status)
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the LinkADRAns payload.
func (cmd *MACCommand_LinkADRAns) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 1 {
		return newLengthUnequalError(1, len(b))
	}
	cmd.ChannelMaskAck = b[0]&1 == 1
	cmd.DataRateIndexAck = (b[0]>>1)&1 == 1
	cmd.TxPowerIndexAck = (b[0]>>2)&1 == 1
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_LinkADRAns_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.LinkADRAns.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_LinkADRAns_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.LinkADRAns.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled DutyCycleReq payload to the slice.
func (cmd *MACCommand_DutyCycleReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if cmd.MaxDutyCycle > 15 {
		return nil, errors.Errorf("expected MaxDutyCycle to be less or equal to 15, got %d", cmd.MaxDutyCycle)
	}
	dst = append(dst, byte(cmd.MaxDutyCycle))
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the DutyCycleReq payload.
func (cmd *MACCommand_DutyCycleReq) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 1 {
		return newLengthUnequalError(1, len(b))
	}
	cmd.MaxDutyCycle = AggregatedDutyCycle(b[0] & 0xf)
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_DutyCycleReq_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.DutyCycleReq.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_DutyCycleReq_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.DutyCycleReq.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled RxParamSetupReq payload to the slice.
func (cmd *MACCommand_RxParamSetupReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if cmd.Rx1DataRateOffset > 7 {
		return nil, errors.Errorf("expected Rx1DROffset to be less or equal to 7, got %d", cmd.Rx1DataRateOffset)
	}
	if cmd.Rx2DataRateIndex > 15 {
		return nil, errors.Errorf("expected Rx2DR to be less or equal to 15, got %d", cmd.Rx2DataRateIndex)
	}
	dst = append(dst, byte(cmd.Rx2DataRateIndex|(cmd.Rx1DataRateOffset<<4)))
	if cmd.Rx2Frequency < 100000 || cmd.Rx2Frequency > maxUint24*100 {
		return nil, errors.Errorf("expected Rx2Frequency to be between %d and %d, got %d", 100000, maxUint24*100, cmd.Rx2Frequency)
	}
	dst = appendUint64(dst, cmd.Rx2Frequency/100, 3)
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the RxParamSetupReq payload.
func (cmd *MACCommand_RxParamSetupReq) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 4 {
		return newLengthUnequalError(4, len(b))
	}
	cmd.Rx1DataRateOffset = uint32((b[0] >> 4) & 0x7)
	cmd.Rx2DataRateIndex = uint32(b[0] & 0xf)
	cmd.Rx2Frequency = parseUint64(b[1:4]) * 100
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_RxParamSetupReq_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.RxParamSetupReq.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_RxParamSetupReq_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.RxParamSetupReq.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled RxParamSetupAns payload to the slice.
func (cmd *MACCommand_RxParamSetupAns) AppendLoRaWAN(dst []byte) ([]byte, error) {
	var b byte
	if cmd.Rx2FrequencyAck {
		b |= 1
	}
	if cmd.Rx2DataRateIndexAck {
		b |= (1 << 1)
	}
	if cmd.Rx1DataRateOffsetAck {
		b |= (1 << 2)
	}
	dst = append(dst, b)
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the RxParamSetupAns payload.
func (cmd *MACCommand_RxParamSetupAns) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 1 {
		return newLengthUnequalError(1, len(b))
	}
	cmd.Rx2FrequencyAck = b[0]&1 == 1
	cmd.Rx2DataRateIndexAck = (b[0]>>1)&1 == 1
	cmd.Rx1DataRateOffsetAck = (b[0]>>2)&1 == 1
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_RxParamSetupAns_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.RxParamSetupAns.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_RxParamSetupAns_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.RxParamSetupAns.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled DevStatusAns payload to the slice.
func (cmd *MACCommand_DevStatusAns) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if cmd.Battery > math.MaxUint8 {
		return nil, errors.Errorf("expected Battery to be less or equal to %d, got %d", math.MaxUint8, cmd.Battery)
	}
	if cmd.Margin < -32 || cmd.Margin > 31 {
		return nil, errors.Errorf("expected Margin to be between -32 and 31, got %d", cmd.Margin)
	}
	dst = append(dst, byte(cmd.Battery))
	if cmd.Margin < 0 {
		dst = append(dst, byte(-(cmd.Margin+1)|(1<<5)))
	} else {
		dst = append(dst, byte(cmd.Margin))
	}
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the DevStatusAns payload.
func (cmd *MACCommand_DevStatusAns) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 2 {
		return newLengthUnequalError(2, len(b))
	}
	cmd.Battery = uint32(b[0])
	margin := int32(b[1] & 0x1f)
	if (b[1]>>5)&1 == 1 {
		margin = -margin - 1
	}
	cmd.Margin = margin
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_DevStatusAns_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.DevStatusAns.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_DevStatusAns_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.DevStatusAns.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled NewChannelReq payload to the slice.
func (cmd *MACCommand_NewChannelReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if cmd.ChannelIndex > math.MaxUint8 {
		return nil, errors.Errorf("expected ChannelIndex to be less or equal to %d, got %d", math.MaxUint8, cmd.ChannelIndex)
	}
	dst = append(dst, byte(cmd.ChannelIndex))

	if cmd.Frequency < 100000 || cmd.Frequency > maxUint24*100 {
		return nil, errors.Errorf("expected Frequency to be between %d and %d, got %d", 100000, maxUint24*100, cmd.Frequency)
	}
	dst = appendUint64(dst, cmd.Frequency/100, 3)

	if cmd.MinDataRateIndex > 15 {
		return nil, errors.Errorf("expected MinDataRateIndex to be less or equal to %d, got %d", 15, cmd.MinDataRateIndex)
	}
	b := byte(cmd.MinDataRateIndex)

	if cmd.MaxDataRateIndex > 15 {
		return nil, errors.Errorf("expected MaxDataRateIndex to be less or equal to %d, got %d", 15, cmd.MaxDataRateIndex)
	}
	b |= byte(cmd.MaxDataRateIndex) << 4
	dst = append(dst, b)
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the NewChannelReq payload.
func (cmd *MACCommand_NewChannelReq) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 5 {
		return newLengthUnequalError(5, len(b))
	}
	cmd.ChannelIndex = uint32(b[0])
	cmd.Frequency = parseUint64(b[1:4]) * 100
	cmd.MinDataRateIndex = uint32(b[4] & 0xf)
	cmd.MaxDataRateIndex = uint32(b[4] >> 4)
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_NewChannelReq_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.NewChannelReq.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_NewChannelReq_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.NewChannelReq.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled NewChannelAns payload to the slice.
func (cmd *MACCommand_NewChannelAns) AppendLoRaWAN(dst []byte) ([]byte, error) {
	var b byte
	if cmd.FrequencyAck {
		b |= 1
	}
	if cmd.DataRateAck {
		b |= (1 << 1)
	}
	dst = append(dst, b)
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the NewChannelAns payload.
func (cmd *MACCommand_NewChannelAns) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 1 {
		return newLengthUnequalError(1, len(b))
	}
	cmd.FrequencyAck = b[0]&1 == 1
	cmd.DataRateAck = (b[0]>>1)&1 == 1
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_NewChannelAns_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.NewChannelAns.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_NewChannelAns_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.NewChannelAns.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled DLChannelReq payload to the slice.
func (cmd *MACCommand_DLChannelReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if cmd.ChannelIndex > math.MaxUint8 {
		return nil, errors.Errorf("expected ChannelIndex to be less or equal to %d, got %d", math.MaxUint8, cmd.ChannelIndex)
	}
	dst = append(dst, byte(cmd.ChannelIndex))

	if cmd.Frequency < 100000 || cmd.Frequency > maxUint24*100 {
		return nil, errors.Errorf("expected Frequency to be between %d and %d, got %d", 100000, maxUint24*100, cmd.Frequency)
	}
	dst = appendUint64(dst, cmd.Frequency/100, 3)
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the DLChannelReq payload.
func (cmd *MACCommand_DLChannelReq) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 4 {
		return newLengthUnequalError(4, len(b))
	}
	cmd.ChannelIndex = uint32(b[0])
	cmd.Frequency = parseUint64(b[1:4]) * 100
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_DlChannelReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.DlChannelReq.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_DlChannelReq) UnmarshalLoRaWAN(b []byte) error {
	return cmd.DlChannelReq.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled DLChannelAns payload to the slice.
func (cmd *MACCommand_DLChannelAns) AppendLoRaWAN(dst []byte) ([]byte, error) {
	var b byte
	if cmd.ChannelIndexAck {
		b |= 1
	}
	if cmd.FrequencyAck {
		b |= (1 << 1)
	}
	dst = append(dst, b)
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the DLChannelAns payload.
func (cmd *MACCommand_DLChannelAns) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 1 {
		return newLengthUnequalError(1, len(b))
	}
	cmd.ChannelIndexAck = b[0]&1 == 1
	cmd.FrequencyAck = (b[0]>>1)&1 == 1
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_DlChannelAns) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.DlChannelAns.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_DlChannelAns) UnmarshalLoRaWAN(b []byte) error {
	return cmd.DlChannelAns.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled RxTimingSetupReq payload to the slice.
func (cmd *MACCommand_RxTimingSetupReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if cmd.Delay > 15 {
		return nil, errors.Errorf("expected Delay to be less or equal to %d, got %d", 15, cmd.Delay)
	}
	dst = append(dst, byte(cmd.Delay))
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the RxTimingSetupReq payload.
func (cmd *MACCommand_RxTimingSetupReq) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 1 {
		return newLengthUnequalError(1, len(b))
	}
	cmd.Delay = uint32(b[0] & 0xf)
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_RxTimingSetupReq_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.RxTimingSetupReq.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_RxTimingSetupReq_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.RxTimingSetupReq.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled TxParamSetupReq payload to the slice.
func (cmd *MACCommand_TxParamSetupReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	var b byte
	if cmd.MaxEIRPIndex > 15 {
		return nil, errors.Errorf("expected MaxEIRPIndex to be less or equal to %d, got %d", 15, cmd.MaxEIRPIndex)
	}
	b = byte(cmd.MaxEIRPIndex)
	if cmd.UplinkDwellTime {
		b |= (1 << 4)
	}
	if cmd.DownlinkDwellTime {
		b |= (1 << 5)
	}
	dst = append(dst, b)
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the TxParamSetupReq payload.
func (cmd *MACCommand_TxParamSetupReq) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 1 {
		return newLengthUnequalError(1, len(b))
	}
	cmd.MaxEIRPIndex = uint32(b[0] & 0xf)
	cmd.UplinkDwellTime = (b[0]>>4)&1 == 1
	cmd.DownlinkDwellTime = (b[0]>>5)&1 == 1
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_TxParamSetupReq_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.TxParamSetupReq.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_TxParamSetupReq_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.TxParamSetupReq.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled RekeyInd payload to the slice.
func (cmd *MACCommand_RekeyInd) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if cmd.MinorVersion > 15 {
		return nil, errors.Errorf("expected MinorVersion to be less or equal to 15, got %d", cmd.MinorVersion)
	}
	dst = append(dst, byte(cmd.MinorVersion))
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the RekeyInd payload.
func (cmd *MACCommand_RekeyInd) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 1 {
		return newLengthUnequalError(1, len(b))
	}
	cmd.MinorVersion = uint32(b[0] & 0xf)
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_RekeyInd_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.RekeyInd.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_RekeyInd_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.RekeyInd.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled RekeyConf payload to the slice.
func (cmd *MACCommand_RekeyConf) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if cmd.MinorVersion > 15 {
		return nil, errors.Errorf("expected MinorVersion to be less or equal to 15, got %d", cmd.MinorVersion)
	}
	dst = append(dst, byte(cmd.MinorVersion))
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the RekeyConf payload.
func (cmd *MACCommand_RekeyConf) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 1 {
		return newLengthUnequalError(1, len(b))
	}
	cmd.MinorVersion = uint32(b[0] & 0xf)
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_RekeyConf_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.RekeyConf.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_RekeyConf_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.RekeyConf.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled ADRParamSetupReq payload to the slice.
func (cmd *MACCommand_ADRParamSetupReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if cmd.ADRAckDelayExponent > 15 {
		return nil, errors.Errorf("expected ADRAckDelayExponent to be less or equal to 15, got %d", cmd.ADRAckDelayExponent)
	}
	b := byte(cmd.ADRAckDelayExponent)

	if cmd.ADRAckLimitExponent > 15 {
		return nil, errors.Errorf("expected ADRAckLimitExponent to be less or equal to 15, got %d", cmd.ADRAckLimitExponent)
	}
	b |= byte(cmd.ADRAckLimitExponent) << 4
	dst = append(dst, b)
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the ADRParamSetupReq payload.
func (cmd *MACCommand_ADRParamSetupReq) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 1 {
		return newLengthUnequalError(1, len(b))
	}
	cmd.ADRAckDelayExponent = uint32(b[0] & 0xf)
	cmd.ADRAckLimitExponent = uint32(b[0] >> 4)
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_ADRParamSetupReq_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.ADRParamSetupReq.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_ADRParamSetupReq_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.ADRParamSetupReq.UnmarshalLoRaWAN(b)
}

// 0.5^8 * 1000000000 ns
const fractStep = 3906250 * time.Nanosecond

// max GPS time allowed in the DeviceTime MAC command
const maxGPSTime int64 = 1<<32 - 1

// AppendLoRaWAN appends the marshaled DeviceTimeAns payload to the slice.
func (cmd *MACCommand_DeviceTimeAns) AppendLoRaWAN(dst []byte) ([]byte, error) {
	sec := gpstime.ToGPS(cmd.Time)
	if sec > maxGPSTime {
		return nil, errors.Errorf("expected GPS time to be less or equal to %d, got %d", maxGPSTime, sec)
	}
	dst = appendUint32(dst, uint32(sec), 4)
	dst = append(dst, byte(time.Duration(cmd.Time.Nanosecond())/fractStep))
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the DeviceTimeAns payload.
func (cmd *MACCommand_DeviceTimeAns) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 5 {
		return newLengthUnequalError(5, len(b))
	}
	cmd.Time = gpstime.Parse(int64(parseUint32(b[0:4])))
	cmd.Time = cmd.Time.Add(time.Duration(b[4]) * fractStep)
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_DeviceTimeAns_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.DeviceTimeAns.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_DeviceTimeAns_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.DeviceTimeAns.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled ForceRejoinReq payload to the slice.
func (cmd *MACCommand_ForceRejoinReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if cmd.PeriodExponent > 7 {
		return nil, errors.Errorf("expected PeriodExponent to be less or equal to 7, got %d", cmd.PeriodExponent)
	}
	// First byte
	b := byte(cmd.PeriodExponent) << 3

	if cmd.MaxRetries > 7 {
		return nil, errors.Errorf("expected MaxRetries to be less or equal to 7, got %d", cmd.MaxRetries)
	}
	b |= byte(cmd.MaxRetries)
	dst = append(dst, b)

	if cmd.RejoinType > 7 {
		return nil, errors.Errorf("expected RejoinType to be less or equal to 7, got %d", cmd.RejoinType)
	}
	// Second byte
	b = byte(cmd.RejoinType) << 4

	if cmd.DataRateIndex > 15 {
		return nil, errors.Errorf("expected DataRateIndex to be less or equal to 15, got %d", cmd.DataRateIndex)
	}
	b |= byte(cmd.DataRateIndex)
	dst = append(dst, b)
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the ForceRejoinReq payload.
func (cmd *MACCommand_ForceRejoinReq) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 2 {
		return newLengthUnequalError(2, len(b))
	}
	cmd.PeriodExponent = uint32(b[0] >> 3)
	cmd.MaxRetries = uint32(b[0] & 0x7)
	cmd.RejoinType = uint32(b[1] >> 4)
	cmd.DataRateIndex = uint32(b[1] & 0xf)
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_ForceRejoinReq_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.ForceRejoinReq.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_ForceRejoinReq_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.ForceRejoinReq.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled RejoinParamSetupReq payload to the slice.
func (cmd *MACCommand_RejoinParamSetupReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if cmd.MaxTimeExponent > 15 {
		return nil, errors.Errorf("expected MaxTimeExponent to be less or equal to 15, got %d", cmd.MaxTimeExponent)
	}
	b := byte(cmd.MaxTimeExponent) << 4

	if cmd.MaxCountExponent > 15 {
		return nil, errors.Errorf("expected MaxCountExponent to be less or equal to 15, got %d", cmd.MaxCountExponent)
	}
	b |= byte(cmd.MaxCountExponent)
	dst = append(dst, b)
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the RejoinParamSetupReq payload.
func (cmd *MACCommand_RejoinParamSetupReq) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 1 {
		return newLengthUnequalError(1, len(b))
	}
	cmd.MaxTimeExponent = uint32(b[0] >> 4)
	cmd.MaxCountExponent = uint32(b[0] & 0xf)
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_RejoinParamSetupReq_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.RejoinParamSetupReq.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_RejoinParamSetupReq_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.RejoinParamSetupReq.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled RejoinParamSetupAns payload to the slice.
func (cmd *MACCommand_RejoinParamSetupAns) AppendLoRaWAN(dst []byte) ([]byte, error) {
	var b byte
	if cmd.MaxTimeExponentAck {
		b |= 1
	}
	dst = append(dst, b)
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the RejoinParamSetupAns payload.
func (cmd *MACCommand_RejoinParamSetupAns) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 1 {
		return newLengthUnequalError(1, len(b))
	}
	cmd.MaxTimeExponentAck = b[0]&1 == 1
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_RejoinParamSetupAns_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.RejoinParamSetupAns.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_RejoinParamSetupAns_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.RejoinParamSetupAns.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled PingSlotInfoReq payload to the slice.
func (cmd *MACCommand_PingSlotInfoReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if cmd.Period > 7 {
		return nil, errors.Errorf("expected Period to be less or equal to 7, got %d", cmd.Period)
	}
	dst = append(dst, byte(cmd.Period))
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the PingSlotInfoReq payload.
func (cmd *MACCommand_PingSlotInfoReq) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 1 {
		return newLengthUnequalError(1, len(b))
	}
	cmd.Period = PingSlotPeriod(b[0] & 0x7)
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_PingSlotInfoReq_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.PingSlotInfoReq.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_PingSlotInfoReq_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.PingSlotInfoReq.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled PingSlotChannelReq payload to the slice.
func (cmd *MACCommand_PingSlotChannelReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if cmd.Frequency > maxUint24 {
		return nil, errors.Errorf("expected Frequency to be less or equal to %d, got %d", maxUint24, cmd.Frequency)
	}
	dst = appendUint64(dst, cmd.Frequency, 3)

	if cmd.DataRateIndex > 15 {
		return nil, errors.Errorf("expected DataRateIndex to be less or equal to 15, got %d", cmd.DataRateIndex)
	}
	dst = append(dst, byte(cmd.DataRateIndex))
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the PingSlotChannelReq payload.
func (cmd *MACCommand_PingSlotChannelReq) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 4 {
		return newLengthUnequalError(4, len(b))
	}
	cmd.Frequency = parseUint64(b[0:3])
	cmd.DataRateIndex = uint32(b[3] & 0xf)
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_PingSlotChannelReq_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.PingSlotChannelReq.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_PingSlotChannelReq_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.PingSlotChannelReq.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled PingSlotChannelAns payload to the slice.
func (cmd *MACCommand_PingSlotChannelAns) AppendLoRaWAN(dst []byte) ([]byte, error) {
	var b byte
	if cmd.FrequencyAck {
		b |= 1
	}
	if cmd.DataRateIndexAck {
		b |= (1 << 1)
	}
	dst = append(dst, b)
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the PingSlotChannelAns payload.
func (cmd *MACCommand_PingSlotChannelAns) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 1 {
		return newLengthUnequalError(1, len(b))
	}
	cmd.FrequencyAck = b[0]&1 == 1
	cmd.DataRateIndexAck = (b[0]>>1)&1 == 1
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_PingSlotChannelAns_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.PingSlotChannelAns.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_PingSlotChannelAns_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.PingSlotChannelAns.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled BeaconTimingAns payload to the slice.
func (cmd *MACCommand_BeaconTimingAns) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if cmd.Delay > math.MaxUint16 {
		return nil, errors.Errorf("expected Delay to be less or equal to %d, got %d", math.MaxUint16, cmd.Delay)
	}
	dst = appendUint32(dst, cmd.Delay, 2)

	if cmd.ChannelIndex > math.MaxUint8 {
		return nil, errors.Errorf("expected ChannelIndex to be less or equal to %d, got %d", math.MaxUint8, cmd.ChannelIndex)
	}
	dst = append(dst, byte(cmd.ChannelIndex))

	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the BeaconTimingAns payload.
func (cmd *MACCommand_BeaconTimingAns) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 3 {
		return newLengthUnequalError(3, len(b))
	}
	cmd.Delay = parseUint32(b[0:2])
	cmd.ChannelIndex = uint32(b[2])
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_BeaconTimingAns_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.BeaconTimingAns.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_BeaconTimingAns_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.BeaconTimingAns.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled BeaconFreqReq payload to the slice.
func (cmd *MACCommand_BeaconFreqReq) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if cmd.Frequency > maxUint24 {
		return nil, errors.Errorf("expected Frequency to be less or equal to %d, got %d", maxUint24, cmd.Frequency)
	}
	dst = appendUint64(dst, cmd.Frequency, 3)
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the BeaconFreqReq payload.
func (cmd *MACCommand_BeaconFreqReq) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 3 {
		return newLengthUnequalError(3, len(b))
	}
	cmd.Frequency = parseUint64(b[0:3])
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_BeaconFreqReq_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.BeaconFreqReq.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_BeaconFreqReq_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.BeaconFreqReq.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled BeaconFreqAns payload to the slice.
func (cmd *MACCommand_BeaconFreqAns) AppendLoRaWAN(dst []byte) ([]byte, error) {
	var b byte
	if cmd.FrequencyAck {
		b |= 1
	}
	dst = append(dst, b)
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the BeaconFreqAns payload.
func (cmd *MACCommand_BeaconFreqAns) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 1 {
		return newLengthUnequalError(1, len(b))
	}
	cmd.FrequencyAck = b[0]&1 == 1
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_BeaconFreqAns_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.BeaconFreqAns.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_BeaconFreqAns_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.BeaconFreqAns.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled DeviceModeInd payload to the slice.
func (cmd *MACCommand_DeviceModeInd) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(cmd.Class))
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the DeviceModeInd payload.
func (cmd *MACCommand_DeviceModeInd) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 1 {
		return newLengthUnequalError(1, len(b))
	}
	cmd.Class = Class(b[0])
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_DeviceModeInd_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.DeviceModeInd.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_DeviceModeInd_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.DeviceModeInd.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled DeviceModeConf payload to the slice.
func (cmd *MACCommand_DeviceModeConf) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, byte(cmd.Class))
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the DeviceModeConf payload.
func (cmd *MACCommand_DeviceModeConf) UnmarshalLoRaWAN(b []byte) error {
	if len(b) != 1 {
		return newLengthUnequalError(1, len(b))
	}
	cmd.Class = Class(b[0])
	return nil
}

// AppendLoRaWAN implements lorawan.Appender.
func (cmd *MACCommand_DeviceModeConf_) AppendLoRaWAN(dst []byte) ([]byte, error) {
	return cmd.DeviceModeConf.AppendLoRaWAN(dst)
}

// UnmarshalLoRaWAN implements lorawan.Unmarshaler.
func (cmd *MACCommand_DeviceModeConf_) UnmarshalLoRaWAN(b []byte) error {
	return cmd.DeviceModeConf.UnmarshalLoRaWAN(b)
}

// AppendLoRaWAN appends the marshaled and payload to the slice.
func (cmd *MACCommand_RawPayload) AppendLoRaWAN(dst []byte) ([]byte, error) {
	dst = append(dst, cmd.RawPayload...)
	return dst, nil
}

// UnmarshalLoRaWAN unmarshals the raw payload.
func (cmd *MACCommand_RawPayload) UnmarshalLoRaWAN(b []byte) error {
	cmd.RawPayload = b
	return nil
}

// AppendLoRaWAN appends the marshaled MAC command and payload to the slice.
func (cmd *MACCommand) AppendLoRaWAN(dst []byte) ([]byte, error) {
	if err := cmd.CID.Validate(); err != nil {
		return nil, err
	}
	dst = append(dst, byte(cmd.CID))

	if cmd.Payload == nil {
		return dst, nil
	}
	return cmd.Payload.(lorawan.Appender).AppendLoRaWAN(dst)
}

// MarshalLoRaWAN marshals the MAC command and payload.
func (cmd *MACCommand) MarshalLoRaWAN() ([]byte, error) {
	// In LoRaWAN1.1 commands contain at most 5 bytes.
	return cmd.AppendLoRaWAN(make([]byte, 0, 5))
}

// UnmarshalLoRaWAN unmarshals the MAC command and payload.
func (cmd *MACCommand) UnmarshalLoRaWAN(b []byte, isUplink bool) error {
	return defaultMACCommands.Read(bytes.NewReader(b), isUplink, cmd)
}

// Read reads a MACCommand from r into cmd and returns any errors encountered.
func (spec macCommandSpec) Read(r io.Reader, isUplink bool, cmd *MACCommand) error {
	b := make([]byte, 1)
	_, err := r.Read(b)
	if err != nil {
		return errors.NewWithCause(err, "failed to read CID")
	}

	ret := MACCommand{
		CID: MACCommandIdentifier(b[0]),
	}

	desc := spec[ret.CID]
	if desc == nil {
		b, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}

		ret.Payload = &MACCommand_RawPayload{
			RawPayload: b,
		}
		*cmd = ret
		return nil
	}

	var pld lorawan.AppendUnmarshaler
	var n uint
	if isUplink {
		n = desc.UplinkLength
		if desc.NewUplink != nil {
			pld = desc.NewUplink()
		}
	} else {
		n = desc.DownlinkLength
		if desc.NewDownlink != nil {
			pld = desc.NewDownlink()
		}
	}

	if n == 0 && pld == nil {
		*cmd = ret
		return nil
	}

	b = make([]byte, n)
	_, err = r.Read(b)
	if err != nil {
		return err
	}

	if pld == nil {
		ret.Payload = &MACCommand_RawPayload{
			RawPayload: b,
		}
		*cmd = ret
		return nil
	}

	switch pld := pld.(type) {
	case isMACCommand_Payload:
		ret.Payload = pld
	case interface {
		MACCommand_Payload() isMACCommand_Payload
	}:
		ret.Payload = pld.MACCommand_Payload()
	default:
		return errors.Errorf("payload has unexpected type: %T", pld)
	}

	if err := pld.UnmarshalLoRaWAN(b); err != nil {
		return errors.NewWithCausef(err, "failed to decode MAC command with CID 0x%X", int32(ret.CID))
	}
	*cmd = ret
	return nil
}

// ReadMACCommand reads a MACCommand from r into cmd and returns any errors encountered.
func ReadMACCommand(r io.Reader, isUplink bool, cmd *MACCommand) error {
	return defaultMACCommands.Read(r, isUplink, cmd)
}
