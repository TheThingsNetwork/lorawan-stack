// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package loraclouddevicemanagementv1

import (
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages/loradms/v1/api/objects"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

var errTLVRecordTooSmall = errors.DefineInvalidArgument("tlv_record_too_small", "TLV record payload is too small")

func parseTLVPayload(record objects.Hex, f func(uint8, uint8, []byte) error) error {
	if len(record) < 2 {
		return errTLVRecordTooSmall.New()
	}
	index := uint8(0)
	for int(index) < len(record) {
		tag := uint8(record[index])
		length := uint8(record[index+1])
		bytes := []byte(record[index+2 : index+2+length])
		index += 2 + length

		if err := f(tag, length, bytes); err != nil {
			return err
		}
	}
	return nil
}
