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

func parseTLVPayload(record objects.Hex, f func(uint8, int, []byte) error) error {
	for len(record) >= 2 {
		tag := record[0]
		length := int(record[1])
		if length+2 > len(record) {
			return errTLVRecordTooSmall.New()
		}

		bytes := []byte(record[2 : 2+length])
		record = record[length+2:]

		if err := f(tag, length, bytes); err != nil {
			return err
		}
	}
	return nil
}
