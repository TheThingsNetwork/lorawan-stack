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
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/mohae/deepcopy"
	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"google.golang.org/grpc"
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

// CopyMACParameters returns a deep copy of ttnpb.MACParameters pb.
func CopyMACParameters(pb *ttnpb.MACParameters) *ttnpb.MACParameters {
	return deepcopy.Copy(pb).(*ttnpb.MACParameters)
}

// CopySessionKeys returns a deep copy of ttnpb.SessionKeys pb.
func CopySessionKeys(pb *ttnpb.SessionKeys) *ttnpb.SessionKeys {
	return deepcopy.Copy(pb).(*ttnpb.SessionKeys)
}

func DurationPtr(v time.Duration) *time.Duration {
	return &v
}

func AES128KeyPtr(key types.AES128Key) *types.AES128Key {
	return &key
}

func MakeEU868Channels(chs ...*ttnpb.MACParameters_Channel) []*ttnpb.MACParameters_Channel {
	return append([]*ttnpb.MACParameters_Channel{
		{
			UplinkFrequency:   868100000,
			DownlinkFrequency: 868100000,
			MinDataRateIndex:  ttnpb.DATA_RATE_0,
			MaxDataRateIndex:  ttnpb.DATA_RATE_5,
		},
		{
			UplinkFrequency:   868300000,
			DownlinkFrequency: 868300000,
			MinDataRateIndex:  ttnpb.DATA_RATE_0,
			MaxDataRateIndex:  ttnpb.DATA_RATE_5,
		},
		{
			UplinkFrequency:   868500000,
			DownlinkFrequency: 868500000,
			MinDataRateIndex:  ttnpb.DATA_RATE_0,
			MaxDataRateIndex:  ttnpb.DATA_RATE_5,
		},
	}, chs...)
}

func NewISPeer(ctx context.Context, is interface {
	ttnpb.ApplicationAccessServer
}) cluster.Peer {
	return test.Must(test.NewGRPCServerPeer(ctx, is, ttnpb.RegisterApplicationAccessServer)).(cluster.Peer)
}

func NewGSPeer(ctx context.Context, gs interface {
	ttnpb.NsGsServer
}) cluster.Peer {
	return test.Must(test.NewGRPCServerPeer(ctx, gs, ttnpb.RegisterNsGsServer)).(cluster.Peer)
}

func NewJSPeer(ctx context.Context, js interface {
	ttnpb.NsJsServer
}) cluster.Peer {
	return test.Must(test.NewGRPCServerPeer(ctx, js, ttnpb.RegisterNsJsServer)).(cluster.Peer)
}

var _ DownlinkTaskQueue = MockDownlinkTaskQueue{}

// MockDownlinkTaskQueue is a mock DownlinkTaskQueue used for testing.
type MockDownlinkTaskQueue struct {
	AddFunc func(ctx context.Context, devID ttnpb.EndDeviceIdentifiers, t time.Time, replace bool) error
	PopFunc func(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error
}

// Add calls AddFunc if set and panics otherwise.
func (m MockDownlinkTaskQueue) Add(ctx context.Context, devID ttnpb.EndDeviceIdentifiers, t time.Time, replace bool) error {
	if m.AddFunc == nil {
		panic("Add called, but not set")
	}
	return m.AddFunc(ctx, devID, t, replace)
}

// Pop calls PopFunc if set and panics otherwise.
func (m MockDownlinkTaskQueue) Pop(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error {
	if m.PopFunc == nil {
		panic("Pop called, but not set")
	}
	return m.PopFunc(ctx, f)
}

type DownlinkTaskAddRequest struct {
	Context     context.Context
	Identifiers ttnpb.EndDeviceIdentifiers
	Time        time.Time
	Replace     bool
	Response    chan<- error
}

func MakeDownlinkTaskAddChFunc(reqCh chan<- DownlinkTaskAddRequest) func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time, bool) error {
	return func(ctx context.Context, devID ttnpb.EndDeviceIdentifiers, t time.Time, replace bool) error {
		respCh := make(chan error)
		reqCh <- DownlinkTaskAddRequest{
			Context:     ctx,
			Identifiers: devID,
			Time:        t,
			Replace:     replace,
			Response:    respCh,
		}
		return <-respCh
	}
}

type DownlinkTaskPopRequest struct {
	Context  context.Context
	Func     func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error
	Response chan<- error
}

func MakeDownlinkTaskPopChFunc(reqCh chan<- DownlinkTaskPopRequest) func(context.Context, func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error {
	return func(ctx context.Context, f func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error {
		respCh := make(chan error)
		reqCh <- DownlinkTaskPopRequest{
			Context:  ctx,
			Func:     f,
			Response: respCh,
		}
		return <-respCh
	}
}

// DownlinkTaskPopBlockFunc is DownlinkTasks.Pop function, which blocks until context is done and returns nil.
func DownlinkTaskPopBlockFunc(ctx context.Context, _ func(context.Context, ttnpb.EndDeviceIdentifiers, time.Time) error) error {
	<-ctx.Done()
	return nil
}

var _ DeviceRegistry = MockDeviceRegistry{}

// MockDeviceRegistry is a mock DeviceRegistry used for testing.
type MockDeviceRegistry struct {
	GetByEUIFunc    func(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, error)
	GetByIDFunc     func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error)
	RangeByAddrFunc func(ctx context.Context, devAddr types.DevAddr, paths []string, f func(*ttnpb.EndDevice) bool) error
	SetByIDFunc     func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
}

// GetByEUI calls GetByEUIFunc if set and panics otherwise.
func (m MockDeviceRegistry) GetByEUI(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, error) {
	if m.GetByEUIFunc == nil {
		panic("GetByEUI called, but not set")
	}
	return m.GetByEUIFunc(ctx, joinEUI, devEUI, paths)
}

// GetByID calls GetByIDFunc if set and panics otherwise.
func (m MockDeviceRegistry) GetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error) {
	if m.GetByIDFunc == nil {
		panic("GetByID called, but not set")
	}
	return m.GetByIDFunc(ctx, appID, devID, paths)
}

// RangeByAddr calls RangeByAddrFunc if set and panics otherwise.
func (m MockDeviceRegistry) RangeByAddr(ctx context.Context, devAddr types.DevAddr, paths []string, f func(*ttnpb.EndDevice) bool) error {
	if m.RangeByAddrFunc == nil {
		panic("RangeByAddr called, but not set")
	}
	return m.RangeByAddrFunc(ctx, devAddr, paths, f)
}

// SetByID calls SetByIDFunc if set and panics otherwise.
func (m MockDeviceRegistry) SetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
	if m.SetByIDFunc == nil {
		panic("SetByID called, but not set")
	}
	return m.SetByIDFunc(ctx, appID, devID, paths, f)
}

type DeviceRegistrySetByIDResponse struct {
	Device *ttnpb.EndDevice
	Error  error
}
type DeviceRegistrySetByIDRequest struct {
	Context                context.Context
	ApplicationIdentifiers ttnpb.ApplicationIdentifiers
	DeviceID               string
	Paths                  []string
	Func                   func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)
	Response               chan<- DeviceRegistrySetByIDResponse
}

func MakeDeviceRegistrySetByIDChFunc(reqCh chan<- DeviceRegistrySetByIDRequest) func(context.Context, ttnpb.ApplicationIdentifiers, string, []string, func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
	return func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
		respCh := make(chan DeviceRegistrySetByIDResponse)
		reqCh <- DeviceRegistrySetByIDRequest{
			Context:                ctx,
			ApplicationIdentifiers: appID,
			DeviceID:               devID,
			Paths:                  paths,
			Func:                   f,
			Response:               respCh,
		}
		resp := <-respCh
		return resp.Device, resp.Error
	}
}

var _ ttnpb.NsJsServer = &MockNsJsServer{}

type MockNsJsServer struct {
	HandleJoinFunc  func(context.Context, *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error)
	GetNwkSKeysFunc func(context.Context, *ttnpb.SessionKeyRequest) (*ttnpb.NwkSKeysResponse, error)
}

// HandleJoin calls HandleJoinFunc if set and panics otherwise.
func (m MockNsJsServer) HandleJoin(ctx context.Context, req *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error) {
	if m.HandleJoinFunc == nil {
		panic("HandleJoin called, but not set")
	}
	return m.HandleJoinFunc(ctx, req)
}

// GetNwkSKeys calls GetNwkSKeysFunc if set and panics otherwise.
func (m MockNsJsServer) GetNwkSKeys(ctx context.Context, req *ttnpb.SessionKeyRequest) (*ttnpb.NwkSKeysResponse, error) {
	if m.GetNwkSKeysFunc == nil {
		panic("GetNwkSKeys called, but not set")
	}
	return m.GetNwkSKeysFunc(ctx, req)
}

type NsJsHandleJoinResponse struct {
	Response *ttnpb.JoinResponse
	Error    error
}
type NsJsHandleJoinRequest struct {
	Context  context.Context
	Message  *ttnpb.JoinRequest
	Response chan<- NsJsHandleJoinResponse
}

func MakeNsJsHandleJoinChFunc(reqCh chan<- NsJsHandleJoinRequest) func(context.Context, *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error) {
	return func(ctx context.Context, msg *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error) {
		respCh := make(chan NsJsHandleJoinResponse)
		reqCh <- NsJsHandleJoinRequest{
			Context:  ctx,
			Message:  msg,
			Response: respCh,
		}
		resp := <-respCh
		return resp.Response, resp.Error
	}
}

var _ ttnpb.NsJsClient = &MockNsJsClient{}

type MockNsJsClient struct {
	*test.MockClientStream
	HandleJoinFunc  func(context.Context, *ttnpb.JoinRequest, ...grpc.CallOption) (*ttnpb.JoinResponse, error)
	GetNwkSKeysFunc func(context.Context, *ttnpb.SessionKeyRequest, ...grpc.CallOption) (*ttnpb.NwkSKeysResponse, error)
}

// HandleJoin calls HandleJoinFunc if set and panics otherwise.
func (m MockNsJsClient) HandleJoin(ctx context.Context, req *ttnpb.JoinRequest, opts ...grpc.CallOption) (*ttnpb.JoinResponse, error) {
	if m.HandleJoinFunc == nil {
		panic("HandleJoin called, but not set")
	}
	return m.HandleJoinFunc(ctx, req, opts...)
}

// GetNwkSKeys calls GetNwkSKeysFunc if set and panics otherwise.
func (m MockNsJsClient) GetNwkSKeys(ctx context.Context, req *ttnpb.SessionKeyRequest, opts ...grpc.CallOption) (*ttnpb.NwkSKeysResponse, error) {
	if m.GetNwkSKeysFunc == nil {
		panic("GetNwkSKeys called, but not set")
	}
	return m.GetNwkSKeysFunc(ctx, req, opts...)
}

var _ ttnpb.NsGsServer = &MockNsGsServer{}

type MockNsGsServer struct {
	ScheduleDownlinkFunc func(context.Context, *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error)
}

// ScheduleDownlink calls ScheduleDownlinkFunc if set and panics otherwise.
func (m MockNsGsServer) ScheduleDownlink(ctx context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
	if m.ScheduleDownlinkFunc == nil {
		panic("ScheduleDownlink called, but not set")
	}
	return m.ScheduleDownlinkFunc(ctx, msg)
}

type NsGsScheduleDownlinkResponse struct {
	Response *ttnpb.ScheduleDownlinkResponse
	Error    error
}
type NsGsScheduleDownlinkRequest struct {
	Context  context.Context
	Message  *ttnpb.DownlinkMessage
	Response chan<- NsGsScheduleDownlinkResponse
}

func MakeNsGsScheduleDownlinkChFunc(reqCh chan<- NsGsScheduleDownlinkRequest) func(context.Context, *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
	return func(ctx context.Context, msg *ttnpb.DownlinkMessage) (*ttnpb.ScheduleDownlinkResponse, error) {
		respCh := make(chan NsGsScheduleDownlinkResponse)
		reqCh <- NsGsScheduleDownlinkRequest{
			Context:  ctx,
			Message:  msg,
			Response: respCh,
		}
		resp := <-respCh
		return resp.Response, resp.Error
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

var _ ttnpb.AsNs_LinkApplicationServer = &MockAsNsLinkApplicationStream{}

type MockAsNsLinkApplicationStream struct {
	*test.MockServerStream
	SendFunc func(*ttnpb.ApplicationUp) error
	RecvFunc func() (*pbtypes.Empty, error)
}

// Send calls SendFunc if set and panics otherwise.
func (m MockAsNsLinkApplicationStream) Send(msg *ttnpb.ApplicationUp) error {
	if m.SendFunc == nil {
		panic("Send called, but not set")
	}
	return m.SendFunc(msg)
}

// Recv calls RecvFunc if set and panics otherwise.
func (m MockAsNsLinkApplicationStream) Recv() (*pbtypes.Empty, error) {
	if m.RecvFunc == nil {
		panic("Recv called, but not set")
	}
	return m.RecvFunc()
}

func AssertDownlinkTaskAddRequest(ctx context.Context, reqCh <-chan DownlinkTaskAddRequest, assert func(ctx context.Context, devID ttnpb.EndDeviceIdentifiers, t time.Time, replace bool) bool, resp error) bool {
	t := test.MustTFromContext(ctx)
	t.Helper()
	select {
	case <-ctx.Done():
		t.Error("Timed out while waiting for DownlinkTasks.Add to be called")
		return false

	case req := <-reqCh:
		if !assert(req.Context, req.Identifiers, req.Time, req.Replace) {
			return false
		}
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for DownlinkTasks.Add response to be processed")
			return false

		case req.Response <- resp:
			return true
		}
	}
}

func AssertNsJsHandleJoinRequest(ctx context.Context, reqCh <-chan NsJsHandleJoinRequest, assert func(ctx context.Context, msg *ttnpb.JoinRequest) bool, resp NsJsHandleJoinResponse) bool {
	t := test.MustTFromContext(ctx)
	t.Helper()
	select {
	case <-ctx.Done():
		t.Error("Timed out while waiting for NsJs.HandleJoin to be called")
		return false

	case req := <-reqCh:
		if !assert(req.Context, req.Message) {
			return false
		}
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for NsJs.HandleJoin response to be processed")
			return false

		case req.Response <- resp:
			return true
		}
	}
}

func AssertAuthNsJsHandleJoinRequest(ctx context.Context, authReqCh <-chan test.ClusterAuthRequest, joinReqCh <-chan NsJsHandleJoinRequest, joinAssert func(ctx context.Context, msg *ttnpb.JoinRequest) bool, authResp grpc.CallOption, joinResp NsJsHandleJoinResponse) bool {
	if !test.AssertClusterAuthRequest(ctx, authReqCh, authResp) {
		return false
	}
	return AssertNsJsHandleJoinRequest(ctx, joinReqCh, joinAssert, joinResp)
}

func AssertNsGsScheduleDownlinkRequest(ctx context.Context, reqCh <-chan NsGsScheduleDownlinkRequest, assert func(ctx context.Context, msg *ttnpb.DownlinkMessage) bool, resp NsGsScheduleDownlinkResponse) bool {
	t := test.MustTFromContext(ctx)
	t.Helper()
	select {
	case <-ctx.Done():
		t.Error("Timed out while waiting for NsGs.ScheduleDownlink to be called")
		return false

	case req := <-reqCh:
		if !assert(req.Context, req.Message) {
			return false
		}
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for NsGs.ScheduleDownlink response to be processed")
			return false

		case req.Response <- resp:
			return true
		}
	}
}

func AssertAuthNsGsScheduleDownlinkRequest(ctx context.Context, authReqCh <-chan test.ClusterAuthRequest, scheduleReqCh <-chan NsGsScheduleDownlinkRequest, scheduleAssert func(ctx context.Context, msg *ttnpb.DownlinkMessage) bool, authResp grpc.CallOption, scheduleResp NsGsScheduleDownlinkResponse) bool {
	if !test.AssertClusterAuthRequest(ctx, authReqCh, authResp) {
		return false
	}
	return AssertNsGsScheduleDownlinkRequest(ctx, scheduleReqCh, scheduleAssert, scheduleResp)
}
