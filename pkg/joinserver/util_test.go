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

package joinserver_test

import (
	"encoding/binary"
	"fmt"
	"strconv"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/joinserver/provisioning"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

var (
	Timeout = (1 << 8) * test.Delay
)

type byteToSerialNumber struct {
}

func (p *byteToSerialNumber) DefaultJoinEUI(entry *pbtypes.Struct) (types.EUI64, error) {
	return types.EUI64{0x42, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, nil
}

func (p *byteToSerialNumber) DefaultDevEUI(entry *pbtypes.Struct) (types.EUI64, error) {
	var devEUI types.EUI64
	binary.BigEndian.PutUint64(devEUI[:], uint64(entry.Fields["serial_number"].GetNumberValue()))
	return devEUI, nil
}

func (p *byteToSerialNumber) DefaultDeviceID(joinEUI, devEUI types.EUI64, entry *pbtypes.Struct) (string, error) {
	return fmt.Sprintf("sn-%d", int(entry.Fields["serial_number"].GetNumberValue())), nil
}

func (p *byteToSerialNumber) UniqueID(entry *pbtypes.Struct) (string, error) {
	return strconv.Itoa(int(entry.Fields["serial_number"].GetNumberValue())), nil
}

func (p *byteToSerialNumber) Decode(data []byte) ([]*pbtypes.Struct, error) {
	var res []*pbtypes.Struct
	for _, b := range data {
		res = append(res, &pbtypes.Struct{
			Fields: map[string]*pbtypes.Value{
				"serial_number": {
					Kind: &pbtypes.Value_NumberValue{
						NumberValue: float64(b),
					},
				},
			},
		})
	}
	return res, nil
}

func init() {
	provisioning.Register("mock", &byteToSerialNumber{})
}
