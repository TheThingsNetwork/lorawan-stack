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

package joinserver

import (
	"context"
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/crypto"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var (
	ErrDevNonceTooSmall     = errDevNonceTooSmall
	ErrInvalidRequest       = errInvalidRequest
	ErrNoAppSKey            = errNoAppSKey
	ErrNoFNwkSIntKey        = errNoFNwkSIntKey
	ErrNoNwkSEncKey         = errNoNwkSEncKey
	ErrNoSessionKeyID       = errNoSessionKeyID
	ErrNoSNwkSIntKey        = errNoSNwkSIntKey
	ErrRegistryOperation    = errRegistryOperation
	ErrReuseDevNonce        = errReuseDevNonce
	ErrSessionKeyIDMismatch = errSessionKeyIDMismatch

	KeyToBytes = keyToBytes
)

type AsJsServer = asJsServer
type NsJsServer = nsJsServer
type JsDeviceServer = jsEndDeviceRegistryServer

type MockDeviceRegistry struct {
	GetByEUIFunc func(context.Context, types.EUI64, types.EUI64, []string) (*ttnpb.EndDevice, error)
	SetByEUIFunc func(context.Context, types.EUI64, types.EUI64, []string, func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
}

func (r *MockDeviceRegistry) GetByEUI(ctx context.Context, joinEUI types.EUI64, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, error) {
	if r.GetByEUIFunc == nil {
		return nil, errors.New("Not implemented")
	}
	return r.GetByEUIFunc(ctx, joinEUI, devEUI, paths)
}

func (r *MockDeviceRegistry) SetByEUI(ctx context.Context, joinEUI types.EUI64, devEUI types.EUI64, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
	if r.SetByEUIFunc == nil {
		return nil, errors.New("Not implemented")
	}
	return r.SetByEUIFunc(ctx, joinEUI, devEUI, paths, f)
}

type MockKeyRegistry struct {
	GetByIDFunc func(context.Context, types.EUI64, []byte, []string) (*ttnpb.SessionKeys, error)
	SetByIDFunc func(context.Context, types.EUI64, []byte, []string, func(*ttnpb.SessionKeys) (*ttnpb.SessionKeys, []string, error)) (*ttnpb.SessionKeys, error)
}

func (r *MockKeyRegistry) GetByID(ctx context.Context, devEUI types.EUI64, id []byte, paths []string) (*ttnpb.SessionKeys, error) {
	if r.GetByIDFunc == nil {
		return nil, errors.New("Not implemented")
	}
	return r.GetByIDFunc(ctx, devEUI, id, paths)
}

func (r *MockKeyRegistry) SetByID(ctx context.Context, devEUI types.EUI64, id []byte, paths []string, f func(*ttnpb.SessionKeys) (*ttnpb.SessionKeys, []string, error)) (*ttnpb.SessionKeys, error) {
	if r.SetByIDFunc == nil {
		return nil, errors.New("Not implemented")
	}
	return r.SetByIDFunc(ctx, devEUI, id, paths, f)
}

func TestMICCheck(t *testing.T) {
	a := assertions.New(t)

	pld := ttnpb.NewPopulatedJoinRequest(test.Randy, false).GetRawPayload()[:19]

	k := types.NewPopulatedAES128Key(test.Randy)
	computed, err := crypto.ComputeJoinRequestMIC(*k, pld)
	if err != nil {
		panic(err)
	}
	a.So(checkMIC(*k, append(pld, computed[:]...)), should.BeNil)
	a.So(checkMIC(*k, append(append(pld[1:], pld[0]-1), computed[:]...)), should.NotBeNil)

	a.So(errors.IsInvalidArgument(checkMIC(*k, []byte{})), should.BeTrue)
}
