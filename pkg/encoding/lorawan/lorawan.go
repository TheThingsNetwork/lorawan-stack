// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

// Package lorawan provides LoRaWAN decoding/encoding interfaces.
package lorawan

import (
	"fmt"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func boolToByte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func copyReverse(dst, src []byte) {
	for i := range src {
		dst[i] = src[len(src)-1-i]
	}
}

func appendReverse(dst []byte, src ...byte) []byte {
	for i := range src {
		dst = append(dst, src[len(src)-1-i])
	}
	return dst
}

// DeviceEIRPToFloat32 returns v as a float32 value.
func DeviceEIRPToFloat32(v ttnpb.DeviceEIRP) float32 {
	switch v {
	case ttnpb.DEVICE_EIRP_36:
		return 36
	case ttnpb.DEVICE_EIRP_33:
		return 33
	case ttnpb.DEVICE_EIRP_30:
		return 30
	case ttnpb.DEVICE_EIRP_29:
		return 29
	case ttnpb.DEVICE_EIRP_27:
		return 27
	case ttnpb.DEVICE_EIRP_26:
		return 26
	case ttnpb.DEVICE_EIRP_24:
		return 24
	case ttnpb.DEVICE_EIRP_21:
		return 21
	case ttnpb.DEVICE_EIRP_20:
		return 20
	case ttnpb.DEVICE_EIRP_18:
		return 18
	case ttnpb.DEVICE_EIRP_16:
		return 16
	case ttnpb.DEVICE_EIRP_14:
		return 14
	case ttnpb.DEVICE_EIRP_13:
		return 13
	case ttnpb.DEVICE_EIRP_12:
		return 12
	case ttnpb.DEVICE_EIRP_10:
		return 10
	case ttnpb.DEVICE_EIRP_8:
		return 8
	}
	panic(fmt.Errorf("unknown DeviceEIRP value `%d`", v))
}

// Float32ToDeviceEIRP returns v as a highest possible DeviceEIRP.
func Float32ToDeviceEIRP(v float32) ttnpb.DeviceEIRP {
	switch {
	case v >= 36:
		return ttnpb.DEVICE_EIRP_36
	case v >= 33:
		return ttnpb.DEVICE_EIRP_33
	case v >= 30:
		return ttnpb.DEVICE_EIRP_30
	case v >= 29:
		return ttnpb.DEVICE_EIRP_29
	case v >= 27:
		return ttnpb.DEVICE_EIRP_27
	case v >= 26:
		return ttnpb.DEVICE_EIRP_26
	case v >= 24:
		return ttnpb.DEVICE_EIRP_24
	case v >= 21:
		return ttnpb.DEVICE_EIRP_21
	case v >= 20:
		return ttnpb.DEVICE_EIRP_20
	case v >= 18:
		return ttnpb.DEVICE_EIRP_18
	case v >= 16:
		return ttnpb.DEVICE_EIRP_16
	case v >= 14:
		return ttnpb.DEVICE_EIRP_14
	case v >= 13:
		return ttnpb.DEVICE_EIRP_13
	case v >= 12:
		return ttnpb.DEVICE_EIRP_12
	case v >= 10:
		return ttnpb.DEVICE_EIRP_10
	}
	return ttnpb.DEVICE_EIRP_8
}

// ADRAckLimitExponentToUint32 returns v as a uint32 value.
func ADRAckLimitExponentToUint32(v ttnpb.ADRAckLimitExponent) uint32 {
	switch v {
	case ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_32768:
		return 32768
	case ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_16384:
		return 16384
	case ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_8192:
		return 8192
	case ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_4096:
		return 4096
	case ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_2048:
		return 2048
	case ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_1024:
		return 1024
	case ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_512:
		return 512
	case ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_256:
		return 256
	case ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_128:
		return 128
	case ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_64:
		return 64
	case ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_32:
		return 32
	case ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_16:
		return 16
	case ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_8:
		return 8
	case ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_4:
		return 4
	case ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_2:
		return 2
	case ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_1:
		return 1
	}
	panic(fmt.Errorf("unknown ADRAckLimitExponent value `%d`", v))
}

// Uint32ToADRAckLimitExponent returns v as a highest possible ADRAckLimitExponent.
func Uint32ToADRAckLimitExponent(v uint32) ttnpb.ADRAckLimitExponent {
	switch {
	case v >= 32768:
		return ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_32768
	case v >= 16384:
		return ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_16384
	case v >= 8192:
		return ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_8192
	case v >= 4096:
		return ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_4096
	case v >= 2048:
		return ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_2048
	case v >= 1024:
		return ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_1024
	case v >= 512:
		return ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_512
	case v >= 256:
		return ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_256
	case v >= 128:
		return ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_128
	case v >= 64:
		return ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_64
	case v >= 32:
		return ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_32
	case v >= 16:
		return ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_16
	case v >= 8:
		return ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_8
	case v >= 4:
		return ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_4
	case v >= 2:
		return ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_2
	}
	return ttnpb.ADRAckLimitExponent_ADR_ACK_LIMIT_1
}

// ADRAckDelayExponentToUint32 returns v as a uint32 value.
func ADRAckDelayExponentToUint32(v ttnpb.ADRAckDelayExponent) uint32 {
	switch v {
	case ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_32768:
		return 32768
	case ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_16384:
		return 16384
	case ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_8192:
		return 8192
	case ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_4096:
		return 4096
	case ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_2048:
		return 2048
	case ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_1024:
		return 1024
	case ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_512:
		return 512
	case ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_256:
		return 256
	case ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_128:
		return 128
	case ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_64:
		return 64
	case ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_32:
		return 32
	case ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_16:
		return 16
	case ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_8:
		return 8
	case ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_4:
		return 4
	case ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_2:
		return 2
	case ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_1:
		return 1
	}
	panic(fmt.Errorf("unknown ADRAckDelayExponent value `%d`", v))
}

// Uint32ToADRAckDelayExponent returns v as a highest possible ADRAckDelayExponent.
func Uint32ToADRAckDelayExponent(v uint32) ttnpb.ADRAckDelayExponent {
	switch {
	case v >= 32768:
		return ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_32768
	case v >= 16384:
		return ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_16384
	case v >= 8192:
		return ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_8192
	case v >= 4096:
		return ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_4096
	case v >= 2048:
		return ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_2048
	case v >= 1024:
		return ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_1024
	case v >= 512:
		return ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_512
	case v >= 256:
		return ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_256
	case v >= 128:
		return ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_128
	case v >= 64:
		return ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_64
	case v >= 32:
		return ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_32
	case v >= 16:
		return ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_16
	case v >= 8:
		return ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_8
	case v >= 4:
		return ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_4
	case v >= 2:
		return ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_2
	}
	return ttnpb.ADRAckDelayExponent_ADR_ACK_DELAY_1
}
