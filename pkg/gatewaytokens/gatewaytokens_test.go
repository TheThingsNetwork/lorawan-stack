// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

package gatewaytokens

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/proto"
)

type mockKS struct{}

var keys = map[string]types.AES128Key{
	"test": {
		0x12, 0x34, 0xAE, 0x00, 0x3A, 0xB7, 0x38, 0x01,
		0x52, 0x31, 0x0B, 0x53, 0x3A, 0xB7, 0x38, 0x01,
	},
}

func (mockKS) HMACHash(_ context.Context, payload []byte, id string) ([]byte, error) {
	key := keys[id]
	h := hmac.New(sha256.New, key[:])
	_, err := h.Write(payload)
	if err != nil {
		return nil, err
	}
	return h.Sum([]byte{}), nil
}

func TestGatewayTokens(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))

	ids := &ttnpb.GatewayIdentifiers{
		GatewayId: "test-gateway",
	}

	rights := &ttnpb.Rights{
		Rights: []ttnpb.Right{
			ttnpb.Right_RIGHT_GATEWAY_LINK,
		},
	}

	mockKV := mockKS{}

	// Generate token.
	token := New(
		"test",
		ids,
		rights,
		mockKV,
	)
	a.So(token, should.NotBeNil)

	// Generate token.
	gen, err := token.Generate(ctx)
	a.So(err, should.BeNil)
	a.So(gen, should.NotBeNil)

	time.Sleep(test.Delay << 2)

	// Expired token
	returnedRights, err := Verify(ctx, gen, test.Delay, mockKV)
	a.So(errors.IsAborted(err), should.BeTrue)
	a.So(returnedRights, should.BeNil)

	// Verify
	returnedRights, err = Verify(ctx, gen, test.Delay<<10, mockKV)
	a.So(err, should.BeNil)
	a.So(returnedRights, should.Resemble, rights)

	// Encoding/Decoding
	msg, err := proto.Marshal(gen)
	a.So(err, should.BeNil)
	a.So(msg, should.NotBeNil)
	decoded, err := DecodeFromString(hex.EncodeToString(msg))
	a.So(err, should.BeNil)
	a.So(decoded, should.Resemble, gen)

	// Multiple Generation
	gen2, err := token.Generate(ctx)
	a.So(err, should.BeNil)
	a.So(gen, should.NotBeNil)
	a.So(gen, should.NotResemble, gen2)
}
