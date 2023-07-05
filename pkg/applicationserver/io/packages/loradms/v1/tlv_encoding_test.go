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
	"fmt"
	"testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages/loradms/v1/api/objects"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestTLVEncoding(t *testing.T) {
	a := assertions.New(t)

	a.So(parseTLVPayload(objects.Hex{0xbb, 0xaa}, func(tag uint8, data []byte) error {
		return fmt.Errorf("foo")
	}), should.NotBeNil)

	a.So(parseTLVPayload(objects.Hex{0x01, 0x02, 0xbb, 0xaa}, func(tag uint8, data []byte) error {
		a.So(tag, should.Equal, 0x01)
		a.So(data, should.HaveLength, 2)
		a.So(data, should.Resemble, []byte{0xbb, 0xaa})
		return nil
	}), should.BeNil)

	a.So(parseTLVPayload(objects.Hex{0xff, 0x02, 0xff}, func(tag uint8, data []byte) error {
		t.Fatal("f should not be called")
		return nil
	}), should.NotBeNil)

	a.So(parseTLVPayload(objects.Hex{0xff, 0xff, 0xff}, func(tag uint8, data []byte) error {
		t.Fatal("f should not be called")
		return nil
	}), should.NotBeNil)
}
