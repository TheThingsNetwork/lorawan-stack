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

// Package crypto implements LoRaWAN crypto.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"fmt"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

var pingOffsetCipher cipher.Block

func init() {
	c, err := aes.NewCipher([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	if err != nil {
		panic(fmt.Sprintf("failed to create ping offset cipher: %s", err))
	}
	pingOffsetCipher = c
}

const (
	minPingPeriod = 1 << 5
	maxPingPeriod = 1 << 12
)

var errInvalidPingPeriod = errors.DefineInvalidArgument("ping_period", fmt.Sprintf("ping period must be a power of 2 between %d and %d, got '{value}'", minPingPeriod, maxPingPeriod))

func ComputePingOffset(beaconTime uint32, devAddr types.DevAddr, pingPeriod uint16) (uint16, error) {
	if pingPeriod < minPingPeriod || pingPeriod > maxPingPeriod {
		return 0, errInvalidPingPeriod.WithAttributes("value", pingPeriod)
	}
	var buf [16]byte
	binary.LittleEndian.PutUint32(buf[0:4], beaconTime)
	copy(buf[4:8], reverse(devAddr[:]))

	var rand [16]byte
	pingOffsetCipher.Encrypt(rand[:], buf[:])

	return (uint16(rand[0]) + uint16(rand[1])*256) % pingPeriod, nil
}
