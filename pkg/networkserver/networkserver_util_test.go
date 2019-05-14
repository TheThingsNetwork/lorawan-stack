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

package networkserver

import (
	"context"
	"errors"
	"testing"
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
	NewMACState = newMACState
	TimePtr     = timePtr

	ErrNoDownlink     = errNoDownlink
	ErrDeviceNotFound = errDeviceNotFound

	FNwkSIntKey = types.AES128Key{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	SNwkSIntKey = types.AES128Key{0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	NwkSEncKey  = types.AES128Key{0x42, 0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	AppSKey     = types.AES128Key{0x42, 0x42, 0x42, 0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

	NetID         = types.NetID{0x42, 0x41, 0x43}
	DevAddr       = types.DevAddr{0x42, 0x42, 0xff, 0xff}
	DevEUI        = types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	JoinEUI       = types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}
	DeviceID      = "test-dev"
	ApplicationID = "test-app"

	Timeout = (1 << 10) * test.Delay
)

// CopyEndDevice returns a deep copy of ttnpb.EndDevice pb.
func CopyEndDevice(pb *ttnpb.EndDevice) *ttnpb.EndDevice {
	return deepcopy.Copy(pb).(*ttnpb.EndDevice)
}

// CopyUplinkMessage returns a deep copy of ttnpb.UplinkMessage pb.
func CopyUplinkMessage(pb *ttnpb.UplinkMessage) *ttnpb.UplinkMessage {
	return deepcopy.Copy(pb).(*ttnpb.UplinkMessage)
}

func DurationPtr(v time.Duration) *time.Duration {
	return &v
}

var _ DownlinkTaskQueue = MockDownlinkTaskQueue{}

// MockDownlinkTaskQueue is a mock DownlinkTaskQueue used for testing.
type MockDownlinkTaskQueue struct {
	AddFunc func(ctx context.Context, devID ttnpb.EndDeviceIdentifiers, t time.Time, replace bool) error
	PopFunc func(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error
}

// Add calls AddFunc if set and returns nil otherwise.
func (q MockDownlinkTaskQueue) Add(ctx context.Context, devID ttnpb.EndDeviceIdentifiers, t time.Time, replace bool) error {
	if q.AddFunc == nil {
		return nil
	}
	return q.AddFunc(ctx, devID, t, replace)
}

// Pop calls PopFunc if set and waits until read on ctx.Done() succeeds and returns ctx.Err() otherwise.
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
	RangeByAddrFunc func(ctx context.Context, devAddr types.DevAddr, paths []string, f func(*ttnpb.EndDevice) bool) error
	SetByIDFunc     func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
}

// GetByEUI calls GetByEUIFunc if set and returns nil, error otherwise.
func (r MockDeviceRegistry) GetByEUI(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, error) {
	if r.GetByEUIFunc == nil {
		return nil, errors.New("GetByEUI must not be called")
	}
	return r.GetByEUIFunc(ctx, joinEUI, devEUI, paths)
}

// GetByID calls GetByIDFunc if set and returns nil, error otherwise.
func (r MockDeviceRegistry) GetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error) {
	if r.GetByIDFunc == nil {
		return nil, errors.New("GetByID must not be called")
	}
	return r.GetByIDFunc(ctx, appID, devID, paths)
}

// RangeByAddr calls RangeByAddrFunc if set and returns error otherwise.
func (r MockDeviceRegistry) RangeByAddr(ctx context.Context, devAddr types.DevAddr, paths []string, f func(*ttnpb.EndDevice) bool) error {
	if r.RangeByAddrFunc == nil {
		return errors.New("RangeByAddr must not be called")
	}
	return r.RangeByAddrFunc(ctx, devAddr, paths, f)
}

// SetByID calls SetByIDFunc if set and returns nil, error otherwise.
func (r MockDeviceRegistry) SetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
	if r.SetByIDFunc == nil {
		return nil, errors.New("SetByID must not be called")
	}
	return r.SetByIDFunc(ctx, appID, devID, paths, f)
}

var _ ttnpb.NsJsServer = &MockNsJsServer{}

type MockNsJsServer struct {
	HandleJoinFunc  func(context.Context, *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error)
	GetNwkSKeysFunc func(context.Context, *ttnpb.SessionKeyRequest) (*ttnpb.NwkSKeysResponse, error)
}

func (js *MockNsJsServer) HandleJoin(ctx context.Context, req *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error) {
	if js.HandleJoinFunc == nil {
		return nil, errors.New("HandleJoin must not be called")
	}
	return js.HandleJoinFunc(ctx, req)
}

func (js *MockNsJsServer) GetNwkSKeys(ctx context.Context, req *ttnpb.SessionKeyRequest) (*ttnpb.NwkSKeysResponse, error) {
	if js.GetNwkSKeysFunc == nil {
		return nil, errors.New("GetNwkSKeys must not be called")
	}
	return js.GetNwkSKeysFunc(ctx, req)
}

type HandleJoinResponse struct {
	Response *ttnpb.JoinResponse
	Error    error
}
type HandleJoinRequest struct {
	Context  context.Context
	Message  *ttnpb.JoinRequest
	Response chan<- HandleJoinResponse
}

func MakeHandleJoinChFunc(reqCh chan<- HandleJoinRequest) func(context.Context, *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error) {
	return func(ctx context.Context, msg *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error) {
		respCh := make(chan HandleJoinResponse)
		reqCh <- HandleJoinRequest{
			Context:  ctx,
			Message:  msg,
			Response: respCh,
		}
		resp := <-respCh
		return resp.Response, resp.Error
	}
}

var _ ttnpb.NsGsServer = &MockNsGsServer{}

type MockNsGsServer struct {
	ScheduleDownlinkFunc func(context.Context, *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error)
}

func (m *MockNsGsServer) ScheduleDownlink(ctx context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
	if m.ScheduleDownlinkFunc == nil {
		return nil, nil
	}
	return m.ScheduleDownlinkFunc(ctx, msg)
}

type ScheduleDownlinkResponse struct {
	Response *ttnpb.ScheduleDownlinkResponse
	Error    error
}
type ScheduleDownlinkRequest struct {
	Context  context.Context
	Message  *ttnpb.DownlinkMessage
	Response chan<- ScheduleDownlinkResponse
}

func MakeScheduleDownlinkChFunc(reqCh chan<- ScheduleDownlinkRequest) func(context.Context, *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
	return func(ctx context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
		respCh := make(chan ScheduleDownlinkResponse)
		reqCh <- ScheduleDownlinkRequest{
			Context:  ctx,
			Message:  msg,
			Response: respCh,
		}
		resp := <-respCh
		return resp.Response, resp.Error
	}
}

func AssertScheduleDownlinkRequest(t *testing.T, reqCh <-chan ScheduleDownlinkRequest, timeout time.Duration, assert func(ctx context.Context, msg *ttnpb.DownlinkMessage) bool, resp ScheduleDownlinkResponse) bool {
	t.Helper()
	select {
	case req := <-reqCh:
		if !assert(req.Context, req.Message) {
			return false
		}
		select {
		case req.Response <- resp:
			return true

		case <-time.After(timeout):
			t.Error("Timed out while waiting for NsGs.ScheduleDownlink response to be processed")
			return false
		}

	case <-time.After(timeout):
		t.Error("Timed out while waiting for NsGs.ScheduleDownlink request to arrive")
		return false
	}
}

type WindowEndRequest struct {
	Context  context.Context
	Message  *ttnpb.UplinkMessage
	Response chan<- time.Time
}

func MakeWindowEndChFunc(reqCh chan<- WindowEndRequest) func(ctx context.Context, msg *ttnpb.UplinkMessage) <-chan time.Time {
	return func(ctx context.Context, msg *ttnpb.UplinkMessage) <-chan time.Time {
		respCh := make(chan time.Time)
		reqCh <- WindowEndRequest{
			Context:  ctx,
			Message:  msg,
			Response: respCh,
		}
		return respCh
	}
}
