// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package networkserver

import (
	"context"
	"errors"
	"time"

	"github.com/mohae/deepcopy"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

const (
	RecentUplinkCount = recentUplinkCount
)

var (
	ResetMACState    = resetMACState
	GenerateDownlink = generateDownlink
	TimePtr          = timePtr

	ErrNoDownlink     = errNoDownlink
	ErrDeviceNotFound = errDeviceNotFound

	FNwkSIntKey = types.AES128Key{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	SNwkSIntKey = types.AES128Key{0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	NwkSEncKey  = types.AES128Key{0x42, 0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	AppSKey     = types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	DevAddr       = types.DevAddr{0x42, 0x42, 0xff, 0xff}
	DevEUI        = types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	JoinEUI       = types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	DeviceID      = "test-dev"
	ApplicationID = "test-app"

	Timeout = (1 << 10) * test.Delay
)

func CopyEndDevice(pb *ttnpb.EndDevice) *ttnpb.EndDevice {
	return deepcopy.Copy(pb).(*ttnpb.EndDevice)
}

func CopyUplinkMessage(pb *ttnpb.UplinkMessage) *ttnpb.UplinkMessage {
	return deepcopy.Copy(pb).(*ttnpb.UplinkMessage)
}

func SetAppQueueUpdateTimeout(d time.Duration) {
	appQueueUpdateTimeout = d
}

var _ DownlinkTaskQueue = MockDownlinkTaskQueue{}

type MockDownlinkTaskQueue struct {
	AddFunc func(ctx context.Context, devID ttnpb.EndDeviceIdentifiers, t time.Time) error
	PopFunc func(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error
}

func (q MockDownlinkTaskQueue) Add(ctx context.Context, devID ttnpb.EndDeviceIdentifiers, t time.Time) error {
	if q.AddFunc == nil {
		return nil
	}
	return q.AddFunc(ctx, devID, t)
}

func (q MockDownlinkTaskQueue) Pop(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error {
	if q.PopFunc == nil {
		<-ctx.Done()
		return ctx.Err()
	}
	return q.PopFunc(ctx, f)
}

var _ DeviceRegistry = MockDeviceRegistry{}

// MockDeviceRegistry is a mock DeviceRegistry used for testing.
type MockDeviceRegistry struct {
	GetByEUIFunc    func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, error)
	GetByIDFunc     func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error)
	RangeByAddrFunc func(devAddr types.DevAddr, paths []string, f func(*ttnpb.EndDevice) bool) error
	SetByIDFunc     func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
}

func (r MockDeviceRegistry) GetByEUI(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, error) {
	if r.GetByEUIFunc == nil {
		return nil, errors.New("GetByEUI not set")
	}
	return r.GetByEUIFunc(ctx, joinEUI, devEUI, paths)
}

func (r MockDeviceRegistry) GetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error) {
	if r.GetByIDFunc == nil {
		return nil, errors.New("GetByID not set")
	}
	return r.GetByIDFunc(ctx, appID, devID, paths)
}

func (r MockDeviceRegistry) RangeByAddr(devAddr types.DevAddr, paths []string, f func(*ttnpb.EndDevice) bool) error {
	if r.RangeByAddrFunc == nil {
		return errors.New("RangeByAddr not set")
	}
	return r.RangeByAddrFunc(devAddr, paths, f)
}

func (r MockDeviceRegistry) SetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
	if r.SetByIDFunc == nil {
		return nil, errors.New("SetByID not set")
	}
	return r.SetByIDFunc(ctx, appID, devID, paths, f)
}
