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
	"context"
	"testing"

	"github.com/smartystreets/assertions"
	clusterauth "go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
	. "go.thethings.network/lorawan-stack/pkg/joinserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestGetAppSKey(t *testing.T) {
	errTest := errors.New("test")

	for _, tc := range []struct {
		Name string

		Context func(context.Context) context.Context

		GetByID     func(context.Context, types.EUI64, []byte, []string) (*ttnpb.SessionKeys, error)
		KeyRequest  *ttnpb.SessionKeyRequest
		KeyResponse *ttnpb.AppSKeyResponse

		ErrorAssertion func(*testing.T, error) bool
	}{
		{
			Name: "Registry error",
			GetByID: func(ctx context.Context, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"app_s_key",
				})
				return nil, errTest
			},
			KeyRequest: &ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: nil,
			ErrorAssertion: func(t *testing.T, err error) bool {
				a := assertions.New(t)
				if !a.So(err, should.EqualErrorOrDefinition, ErrRegistryOperation.WithCause(errTest)) {
					t.FailNow()
				}
				return a.So(errors.IsInternal(err), should.BeTrue)
			},
		},
		{
			Name: "Missing AppSKey",
			GetByID: func(ctx context.Context, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"app_s_key",
				})
				return &ttnpb.SessionKeys{}, nil
			},
			KeyRequest: &ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: nil,
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.EqualErrorOrDefinition, ErrNoAppSKey)
			},
		},
		{
			Name: "Matching request",
			GetByID: func(ctx context.Context, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(devEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(id, should.Resemble, []byte{0x11, 0x22, 0x33, 0x44})
				a.So(paths, should.HaveSameElementsDeep, []string{
					"app_s_key",
				})
				return &ttnpb.SessionKeys{
					SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
					AppSKey: &ttnpb.KeyEnvelope{
						EncryptedKey: KeyToBytes(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
						KEKLabel:     "test-kek",
					},
				}, nil
			},
			KeyRequest: &ttnpb.SessionKeyRequest{
				DevEUI:       types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
			},
			KeyResponse: &ttnpb.AppSKeyResponse{
				AppSKey: ttnpb.KeyEnvelope{
					EncryptedKey: KeyToBytes(types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0x42, 0xff}),
					KEKLabel:     "test-kek",
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			ctx := clusterauth.NewContext(test.ContextWithT(test.Context(), t), nil)
			if tc.Context != nil {
				ctx = tc.Context(ctx)
			}

			js := AsJsServer{
				JS: test.Must(New(
					component.MustNew(test.GetLogger(t), &component.Config{}),
					&Config{
						Keys:    &MockKeyRegistry{GetByIDFunc: tc.GetByID},
						Devices: &MockDeviceRegistry{},
					},
				)).(*JoinServer),
			}
			res, err := js.GetAppSKey(ctx, tc.KeyRequest)

			if tc.ErrorAssertion != nil {
				if !tc.ErrorAssertion(t, err) {
					t.Errorf("Received unexpected error: %s", err)
				}
				a.So(res, should.BeNil)
				return
			}

			a.So(err, should.BeNil)
			a.So(res, should.Resemble, tc.KeyResponse)
		})
	}
}
