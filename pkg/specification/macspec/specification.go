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

// Package macspec provides access to LoRaWAN MAC specification-specific settings.
package macspec

import (
	"fmt"

	"go.thethings.network/lorawan-stack/v3/pkg/crypto"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// macVersion returns the MACVersion as an integer.
func macVersion(v ttnpb.MACVersion) int {
	switch v {
	case ttnpb.MACVersion_MAC_V1_0:
		return 100
	case ttnpb.MACVersion_MAC_V1_0_1:
		return 101
	case ttnpb.MACVersion_MAC_V1_0_2:
		return 102
	case ttnpb.MACVersion_MAC_V1_0_3:
		return 103
	case ttnpb.MACVersion_MAC_V1_0_4:
		return 104
	case ttnpb.MACVersion_MAC_V1_1:
		return 110
	default:
		panic(fmt.Errorf("missed %q in macVersion", v))
	}
}

// compareMACVersion compares MACVersions lhs to rhs:
// -1 == lhs is less than rhs
// 0 == lhs is equal to rhs
// 1 == lhs is greater than rhs
func compareMACVersion(lhs, rhs ttnpb.MACVersion) int {
	switch r := macVersion(lhs) - macVersion(rhs); {
	case r < 0:
		return -1
	case r == 0:
		return 0
	case r > 0:
		return 1
	default:
		panic("unreachable")
	}
}

// EncryptFOpts reports whether v requires MAC commands in FOpts to be encrypted.
func EncryptFOpts(v ttnpb.MACVersion) bool {
	return compareMACVersion(v, ttnpb.MACVersion_MAC_V1_1) >= 0
}

// HasMaxFCntGap reports whether v defines a MaxFCntGap.
func HasMaxFCntGap(v ttnpb.MACVersion) bool {
	return compareMACVersion(v, ttnpb.MACVersion_MAC_V1_0_4) < 0
}

// HasNoChangeADRIndices reports whether v defines a no-change TxPowerIndex and DataRateIndexValue.
func HasNoChangeADRIndices(v ttnpb.MACVersion) bool {
	return compareMACVersion(v, ttnpb.MACVersion_MAC_V1_0_4) >= 0
}

// IgnoreUplinksExceedingLengthLimit reports whether v requires Network Server to
// silently drop uplinks exceeding selected data rate payload length limits.
func IgnoreUplinksExceedingLengthLimit(v ttnpb.MACVersion) bool {
	return compareMACVersion(v, ttnpb.MACVersion_MAC_V1_0_4) >= 0 && compareMACVersion(v, ttnpb.MACVersion_MAC_V1_1) < 0
}

// IncrementDevNonce reports whether v defines DevNonce as an incrementing counter.
func IncrementDevNonce(v ttnpb.MACVersion) bool {
	return compareMACVersion(v, ttnpb.MACVersion_MAC_V1_0_4) >= 0
}

// UseNwkKey reports whether v uses a root NwkKey.
func UseNwkKey(v ttnpb.MACVersion) bool {
	return compareMACVersion(v, ttnpb.MACVersion_MAC_V1_1) >= 0
}

// UseLegacyMIC reports whether v uses legacy MIC computation algorithm.
func UseLegacyMIC(v ttnpb.MACVersion) bool {
	return compareMACVersion(v, ttnpb.MACVersion_MAC_V1_1) < 0
}

// UseSharedFCntDown reports whether v uses the same frame counter
// for both network and application downlinks.
func UseSharedFCntDown(v ttnpb.MACVersion) bool {
	return compareMACVersion(v, ttnpb.MACVersion_MAC_V1_1) < 0
}

// UseDLChannelReq reports whether v is allowed to use the
// DLChannel{Req|Ans} MAC commands.
func UseDLChannelReq(v ttnpb.MACVersion) bool {
	return compareMACVersion(v, ttnpb.MACVersion_MAC_V1_0_2) >= 0
}

// UseTxParamSetupReq reports whether v is allowed to use the
// TxParamSetup{Req|Ans} MAC commands.
func UseTxParamSetupReq(v ttnpb.MACVersion) bool {
	return compareMACVersion(v, ttnpb.MACVersion_MAC_V1_0_2) >= 0
}

// UseADRParamSetupReq reports whether v is allowed to use the
// ADRParamSetup{Req|Ans} MAC commands.
func UseADRParamSetupReq(v ttnpb.MACVersion) bool {
	return compareMACVersion(v, ttnpb.MACVersion_MAC_V1_1) >= 0
}

// UseRejoinParamSetupReq reports whether v is allowed to use the
// RejoinParamSetup{Req|Ans} MAC commands.
func UseRejoinParamSetupReq(v ttnpb.MACVersion) bool {
	return compareMACVersion(v, ttnpb.MACVersion_MAC_V1_1) >= 0
}

// UseDeviceModeInd reports whether v is allowed to use the
// DeviceModeInd MAC command.
func UseDeviceModeInd(v ttnpb.MACVersion) bool {
	return compareMACVersion(v, ttnpb.MACVersion_MAC_V1_1) >= 0
}

// UseRekeyInd reports whether v is allowed to use the
// RekeyInd{Conf} MAC commands.
func UseRekeyInd(v ttnpb.MACVersion) bool {
	return compareMACVersion(v, ttnpb.MACVersion_MAC_V1_1) >= 0
}

// RekeyPeriodVersion returns the version that should be assumed
// after a device joins in order to decrypt the RekeyInd and
// encrypt the RekeyIndConf MAC commands.
func RekeyPeriodVersion(v ttnpb.MACVersion) ttnpb.MACVersion {
	return ttnpb.MACVersion_MAC_V1_1
}

// NegotiatedVersion returns the MAC version and minor version
// that should be used by the end device and network server
// as part of the RekeyInd{Conf} handshake.
// v is the Network Server requested MAC version, while upperBound
// represents the maximum minor accepted by the end device.
func NegotiatedVersion(v ttnpb.MACVersion, upperBound ttnpb.Minor) (ttnpb.MACVersion, ttnpb.Minor) {
	// As there is a singular minor currently available for this mechanism,
	// we return static values.
	return ttnpb.MACVersion_MAC_V1_1, ttnpb.Minor_MINOR_1
}

// AllowDuplicateLinkADRAns reports whether v is allowed to use
// duplicate LinkADRAns MAC responses within the same message.
func AllowDuplicateLinkADRAns(v ttnpb.MACVersion) bool {
	return compareMACVersion(v, ttnpb.MACVersion_MAC_V1_0_2) >= 0 && compareMACVersion(v, ttnpb.MACVersion_MAC_V1_1) < 0
}

// SingularLinkADRAns reports whether v will accept or reject all
// channel masks in a singular LinkADRAns MAC response.
func SingularLinkADRAns(v ttnpb.MACVersion) bool {
	return compareMACVersion(v, ttnpb.MACVersion_MAC_V1_1) >= 0
}

// LimitConfirmedTransmissions reports whether v must limit
// the number of confirmed uplink transmissions.
func LimitConfirmedTransmissions(v ttnpb.MACVersion) bool {
	return compareMACVersion(v, ttnpb.MACVersion_MAC_V1_0_4) < 0
}

// RequireDevEUIForABP reports whether v requires ABP devices to have a DevEUI associated.
func RequireDevEUIForABP(v ttnpb.MACVersion) bool {
	return compareMACVersion(v, ttnpb.MACVersion_MAC_V1_0_4) >= 0
}

// FrameType captures the frame type.
type FrameType uint8

const (
	// UplinkFrame represents the uplink frame.
	UplinkFrame FrameType = iota
	// DownlinkFrame represents the downlink frame.
	DownlinkFrame
)

var (
	fOptsBlockIdentifier1 = [4]byte{0x00, 0x00, 0x00, 0x01}
	fOptsBlockIdentifier2 = [4]byte{0x00, 0x00, 0x00, 0x02}
)

// EncryptionOptions reports the encryption options for v.
func EncryptionOptions(v ttnpb.MACVersion, frameType FrameType, fPort uint32, cmdsInFOpts bool) (opts []crypto.EncryptionOption) {
	// Only FOpts stored as part of the frame header require custom encryption behavior.
	if !cmdsInFOpts {
		return nil
	}
	switch cmp := compareMACVersion(v, ttnpb.MACVersion_MAC_V1_1); {
	case cmp < 0:
		// FOpts are not encrypted before LoRaWAN 1.1.
		return nil
	case cmp >= 0:
		// Do note that the encryption scheme described below is based on the
		// `FOpts Encryption, Usage of FCntDwn Errata on the LoRaWAN L2 1.1 Specification`
		// erratum, and not the literal description found in the LoRaWAN Specification 1.1.
		var (
			isFCntUp    = frameType == UplinkFrame
			isNFCntDown = frameType == DownlinkFrame && fPort == 0
			isAFCntDown = frameType == DownlinkFrame && fPort != 0
		)
		switch {
		case isFCntUp, isNFCntDown:
			return []crypto.EncryptionOption{
				crypto.WithFrameTypeConstant(fOptsBlockIdentifier1),
			}
		case isAFCntDown:
			return []crypto.EncryptionOption{
				crypto.WithFrameTypeConstant(fOptsBlockIdentifier2),
			}
		default:
			panic("unreachable")
		}
	default:
		panic("unreachable")
	}
}

// ValidateUplinkPayloadSize reports whether v requires that the
// Network Server must not force the end device to generate an
// uplink that would be too large for the regional and data rate
// requirements. This can occur if the Network Server includes
// too many MAC commands as part of a downlink.
func ValidateUplinkPayloadSize(v ttnpb.MACVersion) bool {
	return compareMACVersion(v, ttnpb.MACVersion_MAC_V1_1) >= 0
}
