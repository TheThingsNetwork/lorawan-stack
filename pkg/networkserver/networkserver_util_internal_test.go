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
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	clusterauth "go.thethings.network/lorawan-stack/v3/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/frequencyplans"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal"
	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver/internal/test"
	"go.thethings.network/lorawan-stack/v3/pkg/networkserver/mac"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

const (
	DownlinkProcessTaskName          = downlinkProcessTaskName
	DownlinkRetryInterval            = downlinkRetryInterval
	InfrastructureDelay              = infrastructureDelay
	NetworkInitiatedDownlinkInterval = networkInitiatedDownlinkInterval
	RecentDownlinkCount              = recentDownlinkCount
	RecentUplinkCount                = recentUplinkCount
)

var (
	AppendRecentDownlink                = appendRecentDownlink
	AppendRecentUplink                  = appendRecentUplink
	ApplicationJoinAcceptWithoutAppSKey = applicationJoinAcceptWithoutAppSKey
	ApplyCFList                         = applyCFList
	DownlinkPathsFromMetadata           = downlinkPathsFromMetadata
	JoinResponseWithoutKeys             = joinResponseWithoutKeys

	ErrABPJoinRequest             = errABPJoinRequest
	ErrApplicationDownlinkTooLong = errApplicationDownlinkTooLong
	ErrDecodePayload              = errDecodePayload
	ErrDeviceNotFound             = errDeviceNotFound
	ErrDuplicate                  = errDuplicate
	ErrInvalidAbsoluteTime        = errInvalidAbsoluteTime
	ErrOutdatedData               = errOutdatedData
	ErrRejoinRequest              = errRejoinRequest
	ErrUnsupportedLoRaWANVersion  = errUnsupportedLoRaWANVersion

	EvtBeginApplicationLink        = evtBeginApplicationLink
	EvtClusterJoinAttempt          = evtClusterJoinAttempt
	EvtClusterJoinFail             = evtClusterJoinFail
	EvtClusterJoinSuccess          = evtClusterJoinSuccess
	EvtCreateEndDevice             = evtCreateEndDevice
	EvtDropDataUplink              = evtDropDataUplink
	EvtDropJoinRequest             = evtDropJoinRequest
	EvtEndApplicationLink          = evtEndApplicationLink
	EvtForwardDataUplink           = evtForwardDataUplink
	EvtForwardJoinAccept           = evtForwardJoinAccept
	EvtInteropJoinAttempt          = evtInteropJoinAttempt
	EvtInteropJoinFail             = evtInteropJoinFail
	EvtInteropJoinSuccess          = evtInteropJoinSuccess
	EvtProcessDataUplink           = evtProcessDataUplink
	EvtProcessJoinRequest          = evtProcessJoinRequest
	EvtReceiveDataUplink           = evtReceiveDataUplink
	EvtReceiveJoinRequest          = evtReceiveJoinRequest
	EvtScheduleDataDownlinkAttempt = evtScheduleDataDownlinkAttempt
	EvtScheduleDataDownlinkFail    = evtScheduleDataDownlinkFail
	EvtScheduleDataDownlinkSuccess = evtScheduleDataDownlinkSuccess
	EvtScheduleJoinAcceptAttempt   = evtScheduleJoinAcceptAttempt
	EvtScheduleJoinAcceptFail      = evtScheduleJoinAcceptFail
	EvtScheduleJoinAcceptSuccess   = evtScheduleJoinAcceptSuccess
	EvtUpdateEndDevice             = evtUpdateEndDevice

	NewDeviceRegistry         func(t testing.TB) (DeviceRegistry, func())
	NewApplicationUplinkQueue func(t testing.TB) (ApplicationUplinkQueue, func())
	NewDownlinkTaskQueue      func(t testing.TB) (DownlinkTaskQueue, func())
	NewUplinkDeduplicator     func(t testing.TB) (UplinkDeduplicator, func())
)

type DownlinkPath = downlinkPath

var timeMu sync.RWMutex

func SetMockClock(clock *test.MockClock) func() {
	timeMu.Lock()
	oldNow := timeNow
	oldAfter := timeAfter
	timeNow = clock.Now
	timeAfter = clock.After
	return func() {
		timeNow = oldNow
		timeAfter = oldAfter
		timeMu.Unlock()
	}
}

func NSScheduleWindow() time.Duration {
	return nsScheduleWindow()
}

var JoinRequestCorrelationIDs = [...]string{
	"join-request-correlation-id-1",
	"join-request-correlation-id-2",
	"join-request-correlation-id-3",
}

func MakeJoinRequestPHYPayload(joinEUI, devEUI types.EUI64, devNonce types.DevNonce, mic [4]byte) []byte {
	return []byte{
		/* MHDR */
		0b000_000_00,
		joinEUI[7], joinEUI[6], joinEUI[5], joinEUI[4], joinEUI[3], joinEUI[2], joinEUI[1], joinEUI[0],
		devEUI[7], devEUI[6], devEUI[5], devEUI[4], devEUI[3], devEUI[2], devEUI[1], devEUI[0],
		/* DevNonce */
		devNonce[1], devNonce[0],
		/* MIC */
		mic[0], mic[1], mic[2], mic[3],
	}
}

func MakeJoinRequestDecodedPayload(joinEUI, devEUI types.EUI64, devNonce types.DevNonce, mic [4]byte) *ttnpb.Message {
	return &ttnpb.Message{
		MHDR: ttnpb.MHDR{
			MType: ttnpb.MType_JOIN_REQUEST,
			Major: ttnpb.Major_LORAWAN_R1,
		},
		MIC: CopyBytes(mic[:]),
		Payload: &ttnpb.Message_JoinRequestPayload{
			JoinRequestPayload: &ttnpb.JoinRequestPayload{
				JoinEUI:  *joinEUI.Copy(&types.EUI64{}),
				DevEUI:   *devEUI.Copy(&types.EUI64{}),
				DevNonce: deepcopy.Copy(devNonce).(types.DevNonce),
			},
		},
	}
}

type JoinRequestConfig struct {
	DecodePayload bool

	JoinEUI        types.EUI64
	DevEUI         types.EUI64
	DevNonce       types.DevNonce
	DataRate       ttnpb.DataRate
	DataRateIndex  ttnpb.DataRateIndex
	Frequency      uint64
	ChannelIndex   uint8
	ReceivedAt     time.Time
	RxMetadata     []*ttnpb.RxMetadata
	CorrelationIDs []string
	MIC            [4]byte
}

func MakeJoinRequest(conf JoinRequestConfig) *ttnpb.UplinkMessage {
	return MakeUplinkMessage(UplinkMessageConfig{
		RawPayload: MakeJoinRequestPHYPayload(conf.JoinEUI, conf.DevEUI, conf.DevNonce, conf.MIC),
		Payload: func() *ttnpb.Message {
			if conf.DecodePayload {
				return MakeJoinRequestDecodedPayload(conf.JoinEUI, conf.DevEUI, conf.DevNonce, conf.MIC)
			}
			return nil
		}(),
		DataRate:      conf.DataRate,
		DataRateIndex: conf.DataRateIndex,
		Frequency:     conf.Frequency,
		ChannelIndex:  conf.ChannelIndex,
		ReceivedAt:    conf.ReceivedAt,
		RxMetadata:    conf.RxMetadata,
		CorrelationIDs: func() []string {
			if len(conf.CorrelationIDs) == 0 {
				return JoinRequestCorrelationIDs[:]
			}
			return conf.CorrelationIDs
		}(),
	})
}

type NsJsJoinRequestConfig struct {
	JoinEUI            types.EUI64
	DevEUI             types.EUI64
	DevNonce           types.DevNonce
	MIC                [4]byte
	DevAddr            types.DevAddr
	SelectedMACVersion ttnpb.MACVersion
	NetID              types.NetID
	RX1DataRateOffset  uint32
	RX2DataRateIndex   ttnpb.DataRateIndex
	RXDelay            ttnpb.RxDelay
	FrequencyPlanID    string
	PHYVersion         ttnpb.PHYVersion
	CorrelationIDs     []string
}

func MakeNsJsJoinRequest(conf NsJsJoinRequestConfig) *ttnpb.JoinRequest {
	return &ttnpb.JoinRequest{
		RawPayload:         MakeJoinRequestPHYPayload(conf.JoinEUI, conf.DevEUI, conf.DevNonce, conf.MIC),
		Payload:            MakeJoinRequestDecodedPayload(conf.JoinEUI, conf.DevEUI, conf.DevNonce, conf.MIC),
		DevAddr:            *conf.DevAddr.Copy(&types.DevAddr{}),
		SelectedMACVersion: conf.SelectedMACVersion,
		NetID:              *conf.NetID.Copy(&types.NetID{}),
		DownlinkSettings: ttnpb.DLSettings{
			Rx1DROffset: conf.RX1DataRateOffset,
			Rx2DR:       conf.RX2DataRateIndex,
			OptNeg:      conf.SelectedMACVersion.Compare(ttnpb.MAC_V1_1) >= 0,
		},
		RxDelay: conf.RXDelay,
		CFList:  frequencyplans.CFList(*FrequencyPlan(conf.FrequencyPlanID), conf.PHYVersion),
		CorrelationIDs: CopyStrings(func() []string {
			if len(conf.CorrelationIDs) == 0 {
				return JoinRequestCorrelationIDs[:]
			}
			return conf.CorrelationIDs
		}()),
	}
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

var _ InteropClient = MockInteropClient{}

// MockInteropClient is a mock InteropClient used for testing.
type MockInteropClient struct {
	HandleJoinRequestFunc func(context.Context, types.NetID, *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error)
}

// HandleJoinRequest calls HandleJoinRequestFunc if set and panics otherwise.
func (m MockInteropClient) HandleJoinRequest(ctx context.Context, netID types.NetID, req *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error) {
	if m.HandleJoinRequestFunc == nil {
		panic("HandleJoinRequest called, but not set")
	}
	return m.HandleJoinRequestFunc(ctx, netID, req)
}

type InteropClientHandleJoinRequestResponse struct {
	Response *ttnpb.JoinResponse
	Error    error
}

type InteropClientHandleJoinRequestRequest struct {
	Context  context.Context
	NetID    types.NetID
	Request  *ttnpb.JoinRequest
	Response chan<- InteropClientHandleJoinRequestResponse
}

func MakeInteropClientHandleJoinRequestChFunc(reqCh chan<- InteropClientHandleJoinRequestRequest) func(context.Context, types.NetID, *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error) {
	return func(ctx context.Context, netID types.NetID, msg *ttnpb.JoinRequest) (*ttnpb.JoinResponse, error) {
		respCh := make(chan InteropClientHandleJoinRequestResponse)
		reqCh <- InteropClientHandleJoinRequestRequest{
			Context:  ctx,
			NetID:    netID,
			Request:  msg,
			Response: respCh,
		}
		resp := <-respCh
		return resp.Response, resp.Error
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

type InteropClientEnvironment struct {
	HandleJoinRequest <-chan InteropClientHandleJoinRequestRequest
}

func newMockInteropClient(t *testing.T) (InteropClient, InteropClientEnvironment, func()) {
	t.Helper()

	handleJoinCh := make(chan InteropClientHandleJoinRequestRequest)
	return &MockInteropClient{
			HandleJoinRequestFunc: MakeInteropClientHandleJoinRequestChFunc(handleJoinCh),
		}, InteropClientEnvironment{
			HandleJoinRequest: handleJoinCh,
		},
		func() {
			select {
			case <-handleJoinCh:
				t.Error("InteropClient.HandleJoin call missed")
			default:
				close(handleJoinCh)
			}
		}
}

func AssertInteropClientHandleJoinRequestRequest(ctx context.Context, reqCh <-chan InteropClientHandleJoinRequestRequest, assert func(context.Context, types.NetID, *ttnpb.JoinRequest) bool, resp InteropClientHandleJoinRequestResponse) bool {
	t := test.MustTFromContext(ctx)
	t.Helper()
	select {
	case <-ctx.Done():
		t.Error("Timed out while waiting for InteropClient.HandleJoinRequest to be called")
		return false

	case req := <-reqCh:
		if !assert(req.Context, req.NetID, req.Request) {
			return false
		}
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for InteropClient.HandleJoinRequest response to be processed")
			return false

		case req.Response <- resp:
			return true
		}
	}
}

func AssertProcessApplicationUp(ctx context.Context, link ttnpb.AsNs_LinkApplicationClient, assert func(context.Context, *ttnpb.ApplicationUp) bool) bool {
	test.MustTFromContext(ctx).Helper()
	return test.RunSubtestFromContext(ctx, test.SubtestConfig{
		Name: "AsNs.LinkApplication.Recv",
		Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
			t.Helper()

			var asUp *ttnpb.ApplicationUp
			var err error
			if !a.So(test.WaitContext(ctx, func() {
				asUp, err = link.Recv()
			}), should.BeTrue) {
				t.Error("Timed out while waiting for application uplink to be sent to Application Server")
				return
			}
			if !a.So(err, should.BeNil) {
				t.Errorf("Failed to receive Application Server uplink: %s", err)
				return
			}
			if !a.So(assert(ctx, asUp), should.BeTrue) {
				t.Error("Application uplink assertion failed")
				return
			}
			if !a.So(test.WaitContext(ctx, func() {
				err = link.Send(ttnpb.Empty)
			}), should.BeTrue) {
				t.Error("Timed out while waiting for Network Server to process Application Server response")
				return
			}
			if !a.So(err, should.BeNil) {
				t.Errorf("Failed to send Application Server uplink response: %s", err)
				return
			}
		},
	})
}

func AssertProcessApplicationUps(ctx context.Context, link ttnpb.AsNs_LinkApplicationClient, asserts ...func(context.Context, *ttnpb.ApplicationUp) bool) bool {
	t, a := test.MustNewTFromContext(ctx)
	t.Helper()
	for _, assert := range asserts {
		if !a.So(AssertProcessApplicationUp(ctx, link, assert), should.BeTrue) {
			return false
		}
	}
	return !a.Failed()
}

func AssertNetworkServerClose(ctx context.Context, ns *NetworkServer) bool {
	t := test.MustTFromContext(ctx)
	t.Helper()
	if !test.WaitContext(ctx, ns.Close) {
		t.Error("Timed out while waiting for Network Server to gracefully close")
		return false
	}
	return true
}

type TestClusterEnvironment struct {
	Auth    <-chan test.ClusterAuthRequest
	GetPeer <-chan test.ClusterGetPeerRequest
}

type TestEnvironment struct {
	Config

	Cluster       TestClusterEnvironment
	Events        <-chan test.EventPubSubPublishRequest
	InteropClient *InteropClientEnvironment

	*grpc.ClientConn
}

func (env TestEnvironment) AssertLinkApplication(ctx context.Context, appID ttnpb.ApplicationIdentifiers, replaceEvents ...events.Event) (ttnpb.AsNs_LinkApplicationClient, []string, bool) {
	t, a := test.MustNewTFromContext(ctx)
	t.Helper()

	listRightsCh := make(chan test.ApplicationAccessListRightsRequest)
	defer func() {
		close(listRightsCh)
	}()

	var link ttnpb.AsNs_LinkApplicationClient
	var err error
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		link, err = ttnpb.NewAsNsClient(env.ClientConn).LinkApplication(
			(rpcmetadata.MD{
				ID: appID.ApplicationID,
			}).ToOutgoingContext(ctx),
			grpc.PerRPCCredentials(rpcmetadata.MD{
				AuthType:      "Bearer",
				AuthValue:     "link-application-key",
				AllowInsecure: true,
			}),
		)
		wg.Done()
	}()

	var reqCIDs []string
	if !a.So(test.AssertClusterGetPeerRequest(ctx, env.Cluster.GetPeer,
		func(ctx, reqCtx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) (test.ClusterGetPeerResponse, bool) {
			_, a := test.MustNewTFromContext(ctx)
			reqCIDs = events.CorrelationIDsFromContext(reqCtx)
			return test.ClusterGetPeerResponse{
					Peer: NewISPeer(ctx, &test.MockApplicationAccessServer{
						ListRightsFunc: test.MakeApplicationAccessListRightsChFunc(listRightsCh),
					}),
				}, test.AllTrue(
					a.So(reqCIDs, should.NotBeEmpty),
					a.So(role, should.Equal, ttnpb.ClusterRole_ACCESS),
					a.So(ids, should.BeNil),
				)
		},
	), should.BeTrue) {
		return nil, nil, false
	}

	if !a.So(test.AssertListRightsRequest(ctx, listRightsCh,
		func(ctx, reqCtx context.Context, ids ttnpb.Identifiers) bool {
			_, a := test.MustNewTFromContext(ctx)
			md := rpcmetadata.FromIncomingContext(reqCtx)
			cids := events.CorrelationIDsFromContext(reqCtx)
			return a.So(cids, should.NotResemble, reqCIDs) &&
				a.So(cids, should.NotBeEmpty) &&
				a.So(md.AuthType, should.Equal, "Bearer") &&
				a.So(md.AuthValue, should.Equal, "link-application-key") &&
				a.So(ids, should.Resemble, &appID)
		}, ttnpb.RIGHT_APPLICATION_LINK,
	), should.BeTrue) {
		return nil, nil, false
	}

	if !a.So(test.AssertEventPubSubPublishRequests(ctx, env.Events, 1+len(replaceEvents), func(evs ...events.Event) bool {
		return a.So(evs, should.HaveSameElementsEvent, append(
			[]events.Event{EvtBeginApplicationLink.NewWithIdentifiersAndData(events.ContextWithCorrelationID(test.Context(), reqCIDs...), appID, nil)},
			replaceEvents...,
		))
	}), should.BeTrue) {
		t.Error("AS link events assertion failed")
		return nil, nil, false
	}

	if !a.So(test.WaitContext(ctx, wg.Wait), should.BeTrue) {
		t.Error("Timed out while waiting for AS link to open")
		return nil, nil, false
	}
	if !a.So(err, should.BeNil) {
		t.Errorf("Link failed with: %s", err)
		return nil, nil, false
	}
	return link, reqCIDs, a.So(err, should.BeNil)
}

func (env TestEnvironment) AssertWithApplicationLink(ctx context.Context, appID ttnpb.ApplicationIdentifiers, f func(context.Context, ttnpb.AsNs_LinkApplicationClient) bool, replaceEvents ...events.Event) bool {
	t, a := test.MustNewTFromContext(ctx)
	t.Helper()

	ctx, cancel := context.WithCancel(ctx)

	var once sync.Once
	defer once.Do(cancel)

	link, linkCIDs, ok := env.AssertLinkApplication(ctx, appID, replaceEvents...)
	if !test.AllTrue(
		a.So(ok, should.BeTrue),
		f(ctx, link),
	) {
		t.Error("Application link assertion failed")
		return false
	}
	once.Do(cancel)
	if !a.So(env.Events, should.ReceiveEventResembling,
		EvtEndApplicationLink.New(
			events.ContextWithCorrelationID(ctx, linkCIDs...),
			events.WithIdentifiers(appID),
			events.WithData(context.Canceled),
		),
	) {
		t.Error("Link end event assertion failed")
		return false
	}
	return !a.Failed()
}

type DownlinkPathWithPeerIndex struct {
	DownlinkPath
	PeerIndex uint
}

func MakeDownlinkPathsWithPeerIndex(downlinkPaths []DownlinkPath, peerIdxs ...uint) []DownlinkPathWithPeerIndex {
	if len(downlinkPaths) != len(peerIdxs) {
		panic("mismatch in path and index count")
	}
	paths := []DownlinkPathWithPeerIndex{}
	for i, path := range downlinkPaths {
		paths = append(paths, DownlinkPathWithPeerIndex{
			DownlinkPath: path,
			PeerIndex:    peerIdxs[i],
		})
	}
	return paths
}

func UintRepeat(v uint, count int) []uint {
	vs := []uint{}
	for i := 0; i < count; i++ {
		vs = append(vs, v)
	}
	return vs
}

func (env TestEnvironment) AssertLegacyScheduleDownlink(ctx context.Context, paths []DownlinkPathWithPeerIndex, asserts ...func(ctx, reqCtx context.Context, down *ttnpb.DownlinkMessage) (NsGsScheduleDownlinkResponse, bool)) bool {
	test.MustTFromContext(ctx).Helper()
	return test.RunSubtestFromContext(ctx, test.SubtestConfig{
		Name: "NsGs.ScheduleDownlink",
		Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
			t.Helper()

			if len(asserts) > len(paths) {
				panic("invalid assertion count")
			}
			if len(paths) == 0 {
				panic("no paths")
			}

			type Peer struct {
				cluster.Peer
				ScheduleDownlink <-chan NsGsScheduleDownlinkRequest
			}
			peerByIdx := map[uint]Peer{}
			peerByIDs := map[ttnpb.GatewayIdentifiers]Peer{}
			var peerSequence []uint
			for _, path := range paths {
				if path.PeerIndex == 0 {
					continue
				}
				if len(peerSequence) == 0 || peerSequence[len(peerSequence)-1] != path.PeerIndex {
					peerSequence = append(peerSequence, path.PeerIndex)
				}
				peer, ok := peerByIdx[path.PeerIndex]
				if ok {
					peerByIDs[*path.GatewayIdentifiers] = peer
					continue
				}
				scheduleDownlinkCh := make(chan NsGsScheduleDownlinkRequest)
				peer = Peer{
					Peer: NewGSPeer(ctx, &MockNsGsServer{
						ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlinkCh),
					}),
					ScheduleDownlink: scheduleDownlinkCh,
				}
				peerByIdx[path.PeerIndex] = peer
				peerByIDs[*path.GatewayIdentifiers] = peer
			}

			expectedIDs := func() (ids []ttnpb.GatewayIdentifiers) {
				for _, path := range paths {
					ids = append(ids, *path.GatewayIdentifiers)
				}
				return ids
			}()
			var reqIDs []ttnpb.GatewayIdentifiers
			for range paths {
				if !a.So(test.AssertClusterGetPeerRequest(ctx, env.Cluster.GetPeer, func(ctx, reqCtx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) (test.ClusterGetPeerResponse, bool) {
					_, a := test.MustNewTFromContext(ctx)
					gtwIDs, ok := ids.(ttnpb.GatewayIdentifiers)
					if !test.AllTrue(
						a.So(events.CorrelationIDsFromContext(reqCtx), should.NotBeEmpty),
						a.So(role, should.Equal, ttnpb.ClusterRole_GATEWAY_SERVER),
						a.So(ok, should.BeTrue),
					) {
						return test.ClusterGetPeerResponse{
							Error: errors.New("assertion failed"),
						}, false
					}
					if !a.So(expectedIDs, should.Contain, gtwIDs) {
						t.Errorf("Gateway Server peer requested for unknown gateway IDs: %v.\nExpected one of %v", gtwIDs, expectedIDs)
						return test.ClusterGetPeerResponse{
							Error: errors.New("assertion failed"),
						}, false
					}
					reqIDs = append(reqIDs, gtwIDs)
					peer, ok := peerByIDs[gtwIDs]
					if !ok {
						return test.ClusterGetPeerResponse{
							Error: errPeerNotFound.New(),
						}, true
					}
					return test.ClusterGetPeerResponse{
						Peer: peer,
					}, true
				}), should.BeTrue) {
					t.Error("Gateway Server peer look-up assertion failed")
					return
				}
			}
			if !a.So(reqIDs, should.HaveSameElementsDeep, expectedIDs) {
				t.Errorf("Gateway peers by incorrect gateway IDs were requested: %v.\nExpected peers for following gateway IDs to be requested: %v", reqIDs, expectedIDs)
			}

			if len(asserts) > len(peerSequence) {
				panic(fmt.Errorf("mismatch in assertion count and ScheduleDownlink calls: %d assertions, %d ScheduleDownlink calls; peer sequence: %v", len(asserts), len(peerSequence), peerSequence))
			}

			for i, assert := range asserts {
				if !a.So(test.AssertClusterAuthRequest(
					ctx,
					env.Cluster.Auth,
					&grpc.EmptyCallOption{},
				), should.BeTrue) {
					t.Errorf("Failed to assert Cluster.Auth request for schedule attempt number %d", i)
					return
				}
				select {
				case <-ctx.Done():
					t.Errorf("Timed out while waiting for NsGs.ScheduleDownlink to be called for schedule attempt number %d", i)
					return
				case req := <-peerByIdx[peerSequence[i]].ScheduleDownlink:
					resp, ok := assert(ctx, req.Context, req.Message)
					if !a.So(ok, should.BeTrue) {
						t.Errorf("NsGs.ScheduleDownlink request assertion failed for schedule attempt number %d", i)
						return
					}
					select {
					case <-ctx.Done():
						t.Errorf("Timed out while waiting for NsGs.ScheduleDownlink response to be processed for schedule attempt number %d", i)
						return

					case req.Response <- resp:
					}
				}
			}
		},
	})
}

var errPeerNotFound = errors.DefineNotFound("test_peer", "test peer not found")

type DownlinkSchedulingAssertionConfig struct {
	SetRX1          bool
	SetRX2          bool
	FrequencyPlanID string
	PHYVersion      ttnpb.PHYVersion
	MACState        *ttnpb.MACState
	Session         *ttnpb.Session
	Class           ttnpb.Class
	RX1Delay        ttnpb.RxDelay
	Uplink          *ttnpb.UplinkMessage
	Priority        ttnpb.TxSchedulePriority
	AbsoluteTime    *time.Time
	FixedPaths      []ttnpb.GatewayAntennaIdentifiers
	Payload         []byte
	CorrelationIDs  []string
	PeerIndexes     []uint
	Responses       []NsGsScheduleDownlinkResponse
}

func (env TestEnvironment) AssertScheduleDownlink(ctx context.Context, conf DownlinkSchedulingAssertionConfig) (*ttnpb.DownlinkMessage, bool) {
	t, a := test.MustNewTFromContext(ctx)
	t.Helper()
	var lastDown *ttnpb.DownlinkMessage
	return lastDown, a.So(test.RunSubtestFromContext(ctx, test.SubtestConfig{
		Name: "NsGs.ScheduleDownlink",
		Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
			t.Helper()

			fp := FrequencyPlan(conf.FrequencyPlanID)
			phy := LoRaWANBands[fp.BandID][conf.PHYVersion]

			var downlinkPaths []DownlinkPath
			if conf.Uplink != nil {
				downlinkPaths = DownlinkPathsFromMetadata(conf.Uplink.RxMetadata...)
			} else {
				for i := range conf.FixedPaths {
					downlinkPaths = append(downlinkPaths, DownlinkPath{
						GatewayIdentifiers: &conf.FixedPaths[i].GatewayIdentifiers,
						DownlinkPath: &ttnpb.DownlinkPath{
							Path: &ttnpb.DownlinkPath_Fixed{
								Fixed: &conf.FixedPaths[i],
							},
						},
					})
				}
			}
			if len(downlinkPaths) == 0 {
				panic("no paths")
			}

			type Peer struct {
				cluster.Peer
				ScheduleDownlink <-chan NsGsScheduleDownlinkRequest
			}
			type ExpectedAttempt struct {
				PeerIndex    uint
				RequestPaths []*ttnpb.DownlinkPath
			}
			peerByIdx := map[uint]*Peer{}
			peerByIDs := map[ttnpb.GatewayIdentifiers]*Peer{}
			var expectedAttempts []ExpectedAttempt
			for i, path := range downlinkPaths {
				if len(conf.PeerIndexes) <= i || conf.PeerIndexes[i] == 0 {
					continue
				}
				peer, ok := peerByIdx[conf.PeerIndexes[i]]
				if !ok {
					scheduleDownlinkCh := make(chan NsGsScheduleDownlinkRequest)
					peer = &Peer{
						Peer: NewGSPeer(ctx, &MockNsGsServer{
							ScheduleDownlinkFunc: MakeNsGsScheduleDownlinkChFunc(scheduleDownlinkCh),
						}),
						ScheduleDownlink: scheduleDownlinkCh,
					}
					peerByIdx[conf.PeerIndexes[i]] = peer
				}
				peerByIDs[*path.GatewayIdentifiers] = peer
				if len(expectedAttempts) == 0 || expectedAttempts[len(expectedAttempts)-1].PeerIndex != conf.PeerIndexes[i] {
					expectedAttempts = append(expectedAttempts, ExpectedAttempt{
						PeerIndex:    conf.PeerIndexes[i],
						RequestPaths: []*ttnpb.DownlinkPath{path.DownlinkPath},
					})
				} else {
					n := len(expectedAttempts)
					expectedAttempts[n].RequestPaths = append(expectedAttempts[n].RequestPaths, path.DownlinkPath)
				}
			}

			expectedIDs := func() (ids []ttnpb.GatewayIdentifiers) {
				for _, path := range downlinkPaths {
					ids = append(ids, *path.GatewayIdentifiers)
				}
				return ids
			}()
			var reqIDs []ttnpb.GatewayIdentifiers
			for range downlinkPaths {
				if !a.So(test.AssertClusterGetPeerRequest(ctx, env.Cluster.GetPeer, func(ctx, reqCtx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) (test.ClusterGetPeerResponse, bool) {
					_, a := test.MustNewTFromContext(ctx)
					gtwIDs, ok := ids.(ttnpb.GatewayIdentifiers)
					if !test.AllTrue(
						a.So(events.CorrelationIDsFromContext(reqCtx), should.NotBeEmpty),
						a.So(role, should.Equal, ttnpb.ClusterRole_GATEWAY_SERVER),
						a.So(ok, should.BeTrue),
					) {
						return test.ClusterGetPeerResponse{
							Error: errors.New("assertion failed"),
						}, false
					}
					if !a.So(expectedIDs, should.Contain, gtwIDs) {
						t.Errorf("Gateway Server peer requested for unknown gateway IDs: %v.\nExpected one of %v", gtwIDs, expectedIDs)
						return test.ClusterGetPeerResponse{
							Error: errors.New("assertion failed"),
						}, false
					}
					reqIDs = append(reqIDs, gtwIDs)
					peer, ok := peerByIDs[gtwIDs]
					if !ok {
						return test.ClusterGetPeerResponse{
							Error: errPeerNotFound.New(),
						}, true
					}
					return test.ClusterGetPeerResponse{
						Peer: peer,
					}, true
				}), should.BeTrue) {
					t.Error("Gateway Server peer look-up assertion failed")
					return
				}
			}
			if !a.So(reqIDs, should.HaveSameElementsDeep, expectedIDs) {
				t.Errorf("Gateway peers by incorrect gateway IDs were requested: %v.\nExpected peers for following gateway IDs to be requested: %v", reqIDs, expectedIDs)
			}

			if len(conf.Responses) > len(expectedAttempts) {
				panic(fmt.Errorf("mismatch in response count and expected attempt count: %d responses, %d expected attempts; expected attempts: %v", len(conf.Responses), len(expectedAttempts), expectedAttempts))
			}

			expectedCIDs := conf.CorrelationIDs
			if conf.Uplink != nil {
				expectedCIDs = append(expectedCIDs, conf.Uplink.CorrelationIDs...)
			}
			for i, expectedAttempt := range expectedAttempts {
				if !a.So(test.AssertClusterAuthRequest(
					ctx,
					env.Cluster.Auth,
					&grpc.EmptyCallOption{},
				), should.BeTrue) {
					t.Errorf("Failed to assert Cluster.Auth request for schedule attempt number %d", i)
					return
				}
				select {
				case <-ctx.Done():
					t.Errorf("Timed out while waiting for NsGs.ScheduleDownlink to be called for schedule attempt number %d", i)
					return
				case req := <-peerByIdx[expectedAttempt.PeerIndex].ScheduleDownlink:
					lastDown = req.Message

					if !test.AllTrue(
						a.So(req.Message.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual, expectedCIDs),
						a.So(req.Message, should.Resemble, &ttnpb.DownlinkMessage{
							RawPayload: conf.Payload,
							Settings: &ttnpb.DownlinkMessage_Request{
								Request: func() *ttnpb.TxRequest {
									txReq := &ttnpb.TxRequest{
										Class:           conf.Class,
										DownlinkPaths:   expectedAttempt.RequestPaths,
										Priority:        conf.Priority,
										FrequencyPlanID: conf.FrequencyPlanID,
										AbsoluteTime:    conf.AbsoluteTime,
									}
									if conf.SetRX1 {
										txReq.Rx1Delay = conf.RX1Delay
										txReq.Rx1DataRateIndex = test.Must(phy.Rx1DataRate(
											conf.Uplink.Settings.DataRateIndex,
											conf.MACState.CurrentParameters.Rx1DataRateOffset,
											conf.MACState.CurrentParameters.DownlinkDwellTime.GetValue()),
										).(ttnpb.DataRateIndex)
										txReq.Rx1Frequency = conf.MACState.CurrentParameters.Channels[test.Must(phy.Rx1Channel(uint8(conf.Uplink.DeviceChannelIndex))).(uint8)].DownlinkFrequency
									}
									if conf.SetRX2 {
										txReq.Rx2DataRateIndex = conf.MACState.CurrentParameters.Rx2DataRateIndex
										txReq.Rx2Frequency = conf.MACState.CurrentParameters.Rx2Frequency
									}
									return txReq
								}(),
							},
							CorrelationIDs: req.Message.CorrelationIDs,
						}),
					) {
						t.Errorf("NsGs.ScheduleDownlink request assertion failed for schedule attempt number %d", i)
						return
					}
					select {
					case <-ctx.Done():
						t.Errorf("Timed out while waiting for NsGs.ScheduleDownlink response to be processed for schedule attempt number %d", i)
						return

					case req.Response <- conf.Responses[i]:
					}
				}
			}
		},
	}), should.BeTrue)
}

func (env TestEnvironment) AssertScheduleJoinAccept(ctx context.Context, dev *ttnpb.EndDevice) (*ttnpb.EndDevice, bool) {
	t, a := test.MustNewTFromContext(ctx)
	t.Helper()
	dev = CopyEndDevice(dev)
	return dev, a.So(test.RunSubtestFromContext(ctx, test.SubtestConfig{
		Name: "Join-accept",
		Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
			t.Helper()

			fp := FrequencyPlan(dev.FrequencyPlanID)
			phy := LoRaWANBands[fp.BandID][dev.LoRaWANPHYVersion]

			scheduledDown, ok := env.AssertScheduleDownlink(ctx, DownlinkSchedulingAssertionConfig{
				SetRX1:          true,
				SetRX2:          true,
				FrequencyPlanID: dev.FrequencyPlanID,
				PHYVersion:      dev.LoRaWANPHYVersion,
				MACState:        dev.PendingMACState,
				Session:         dev.PendingSession,
				Class:           ttnpb.CLASS_A,
				RX1Delay:        ttnpb.RxDelay(phy.JoinAcceptDelay1.Seconds()),
				Uplink:          LastUplink(dev.RecentUplinks...),
				Priority:        ttnpb.TxSchedulePriority_HIGHEST,
				Payload:         dev.PendingMACState.QueuedJoinAccept.Payload,
				CorrelationIDs:  dev.PendingMACState.QueuedJoinAccept.CorrelationIDs,
				PeerIndexes:     []uint{1},
				Responses: []NsGsScheduleDownlinkResponse{
					{
						Response: &ttnpb.ScheduleDownlinkResponse{},
					},
				},
			})
			if !a.So(ok, should.BeTrue) {
				t.Error("Join-accept scheduling assertion failed")
				return
			}
			a.So(env.Events, should.ReceiveEventsResembling,
				EvtScheduleJoinAcceptAttempt.With(
					events.WithData(&ttnpb.DownlinkMessage{
						RawPayload: dev.PendingMACState.QueuedJoinAccept.Payload,
						Payload: &ttnpb.Message{
							MHDR: ttnpb.MHDR{
								MType: ttnpb.MType_JOIN_ACCEPT,
								Major: ttnpb.Major_LORAWAN_R1,
							},
							Payload: &ttnpb.Message_JoinAcceptPayload{
								JoinAcceptPayload: &ttnpb.JoinAcceptPayload{
									NetID:      dev.PendingMACState.QueuedJoinAccept.Request.NetID,
									DevAddr:    dev.PendingMACState.QueuedJoinAccept.Request.DevAddr,
									DLSettings: dev.PendingMACState.QueuedJoinAccept.Request.DownlinkSettings,
									RxDelay:    dev.PendingMACState.QueuedJoinAccept.Request.RxDelay,
									CFList:     dev.PendingMACState.QueuedJoinAccept.Request.CFList,
								},
							},
						},
						Settings:       scheduledDown.Settings,
						CorrelationIDs: scheduledDown.CorrelationIDs,
					}),
					events.WithIdentifiers(dev.EndDeviceIdentifiers),
				).New(ctx),
				EvtScheduleJoinAcceptSuccess.With(
					events.WithData(&ttnpb.ScheduleDownlinkResponse{}),
					events.WithIdentifiers(dev.EndDeviceIdentifiers),
				).New(events.ContextWithCorrelationID(ctx, scheduledDown.CorrelationIDs...)),
			)
			dev.PendingSession = &ttnpb.Session{
				DevAddr:     dev.PendingMACState.QueuedJoinAccept.Request.DevAddr,
				SessionKeys: dev.PendingMACState.QueuedJoinAccept.Keys,
			}
			dev.PendingMACState.PendingJoinRequest = &dev.PendingMACState.QueuedJoinAccept.Request
			dev.PendingMACState.QueuedJoinAccept = nil
			dev.PendingMACState.RxWindowsAvailable = false
			dev.RecentDownlinks = AppendRecentDownlink(dev.RecentDownlinks, scheduledDown, RecentDownlinkCount)
		},
	}), should.BeTrue)
}

type DataDownlinkAssertionConfig struct {
	SetRX1         bool
	SetRX2         bool
	Device         *ttnpb.EndDevice
	Class          ttnpb.Class
	Priority       ttnpb.TxSchedulePriority
	AbsoluteTime   *time.Time
	FixedPaths     []ttnpb.GatewayAntennaIdentifiers
	RawPayload     []byte
	Payload        *ttnpb.Message
	CorrelationIDs []string
	PeerIndexes    []uint
	Responses      []NsGsScheduleDownlinkResponse
}

func (env TestEnvironment) AssertScheduleDataDownlink(ctx context.Context, conf DataDownlinkAssertionConfig) (*ttnpb.EndDevice, bool) {
	t, a := test.MustNewTFromContext(ctx)
	t.Helper()
	dev := CopyEndDevice(conf.Device)
	return dev, a.So(test.RunSubtestFromContext(ctx, test.SubtestConfig{
		Name: "Data downlink",
		Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
			t.Helper()

			scheduledDown, ok := env.AssertScheduleDownlink(ctx, DownlinkSchedulingAssertionConfig{
				SetRX1:          conf.SetRX1,
				SetRX2:          conf.SetRX2,
				FrequencyPlanID: dev.FrequencyPlanID,
				PHYVersion:      dev.LoRaWANPHYVersion,
				MACState:        dev.MACState,
				Session:         dev.Session,
				Class:           conf.Class,
				RX1Delay:        dev.MACState.CurrentParameters.Rx1Delay,
				Uplink:          LastUplink(dev.MACState.RecentUplinks...),
				Priority:        conf.Priority,
				Payload:         conf.RawPayload,
				PeerIndexes:     conf.PeerIndexes,
				Responses:       conf.Responses,
			})
			a.So(ok, should.BeTrue)
			a.So(env.Events, should.ReceiveEventsResembling,
				EvtScheduleDataDownlinkAttempt.With(
					events.WithData(&ttnpb.DownlinkMessage{
						RawPayload:     conf.RawPayload,
						Payload:        conf.Payload,
						Settings:       scheduledDown.Settings,
						CorrelationIDs: scheduledDown.CorrelationIDs,
					}),
					events.WithIdentifiers(dev.EndDeviceIdentifiers),
				).New(ctx),
				EvtScheduleDataDownlinkSuccess.With(
					events.WithData(&ttnpb.ScheduleDownlinkResponse{}),
					events.WithIdentifiers(dev.EndDeviceIdentifiers),
				).New(events.ContextWithCorrelationID(ctx, scheduledDown.CorrelationIDs...)),
			)
			dev.RecentDownlinks = AppendRecentDownlink(dev.RecentDownlinks, scheduledDown, RecentDownlinkCount)
		},
	}), should.BeTrue)
}

func (env TestEnvironment) AssertHandleDeviceUplink(ctx context.Context, assert func(context.Context, func(...events.Event) bool) (func(context.Context, error) bool, bool), ups ...*ttnpb.UplinkMessage) bool {
	t, a := test.MustNewTFromContext(ctx)
	t.Helper()
	return a.So(test.RunSubtestFromContext(ctx, test.SubtestConfig{
		Name: "GsNs.HandleUplink",
		Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
			t.Helper()

			errCh := make(chan error, len(ups))
			wg := &sync.WaitGroup{}
			wg.Add(len(ups) - 1)
			go func() {
				t.Logf("Call GsNs.HandleUplink with first uplink: %v", ups[0])
				_, err := ttnpb.NewGsNsClient(env.ClientConn).HandleUplink(ctx, ups[0])
				t.Logf("First GsNs.HandleUplink returned %v", err)
				errCh <- err
				wg.Wait()
				close(errCh)
			}()
			for _, up := range ups[1:] {
				up := up
				time.AfterFunc(env.Config.DeduplicationWindow/2, func() {
					t.Logf("Call GsNs.HandleUplink with duplicate uplink: %v", up)
					_, err := ttnpb.NewGsNsClient(env.ClientConn).HandleUplink(ctx, up)
					t.Logf("Duplicate GsNs.HandleUplink returned %v", err)
					errCh <- err
					wg.Done()
				})
			}
			assertError, ok := assert(ctx, func(expectedEvs ...events.Event) bool {
				return a.So(test.RunSubtestFromContext(ctx, test.SubtestConfig{
					Name: "uplink handling events",
					Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
						t.Helper()

						a.So(test.AssertEventPubSubPublishRequests(ctx, env.Events, len(expectedEvs), func(evs ...events.Event) bool {
							if !a.So(evs, should.HaveSameElementsFunc, test.MakeEventEqual(test.EventEqualConfig{
								Identifiers:    true,
								Origin:         true,
								Context:        true,
								Visibility:     true,
								Authentication: true,
								RemoteIP:       true,
								UserAgent:      true,
							}), expectedEvs) {
								printEvents := func(evs []events.Event) string {
									var s string
									for i, ev := range evs {
										s += fmt.Sprintf("\nevent %d: %s", i, ev)
									}
									return s
								}
								t.Errorf("Uplink event assertion failed.\nGot events: %s\nExpected events: %s", printEvents(evs), printEvents(expectedEvs))
								return false
							}
							return true
						}), should.BeTrue)
					},
				}), should.BeTrue)
			})
			if !a.So(ok, should.BeTrue) {
				t.Error("Uplink handling assertion failed")
				return
			}
			for range ups[1:] {
				select {
				case <-ctx.Done():
					t.Error("Timed out while waiting for duplicate HandleUplink to return")
					return

				case err := <-errCh:
					if !a.So(err, should.BeNil) {
						t.Errorf("Failed to handle duplicate uplink: %s", err)
						return
					}
				}
			}
			select {
			case <-ctx.Done():
				t.Error("Timed out while waiting for HandleUplink to return")
				return
			case err := <-errCh:
				var ok bool
				if assertError == nil {
					ok = a.So(err, should.BeNil)
				} else {
					ok = a.So(assertError(ctx, err), should.BeTrue)
				}
				if !ok {
					t.Errorf("HandleUplink error assertion failed")
					return
				}
			}
		},
	}), should.BeTrue)
}

func (env TestEnvironment) AssertHandleDeviceUplinkSuccess(ctx context.Context, assert func(context.Context, func(...events.Event) bool) bool, ups ...*ttnpb.UplinkMessage) bool {
	t, a := test.MustNewTFromContext(ctx)
	t.Helper()
	return a.So(env.AssertHandleDeviceUplink(
		ctx,
		func(ctx context.Context, assertEvents func(...events.Event) bool) (func(context.Context, error) bool, bool) {
			_, a := test.MustNewTFromContext(ctx)
			return nil, a.So(assert(ctx, assertEvents), should.BeTrue)
		},
		ups...,
	), should.BeTrue)
}

func (env TestEnvironment) AssertHandleJoinRequest(ctx context.Context, conf JoinRequestConfig, assert func(ctx context.Context, assertEvents func(...events.Event) bool, ups ...*ttnpb.UplinkMessage) bool, duplicateMDs ...[]*ttnpb.RxMetadata) bool {
	t, a := test.MustNewTFromContext(ctx)
	t.Helper()
	return a.So(test.RunSubtestFromContext(ctx, test.SubtestConfig{
		Name: "Join-request",
		Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
			t.Helper()

			ups := []*ttnpb.UplinkMessage{MakeJoinRequest(conf)}
			for _, mds := range duplicateMDs {
				mds := mds
				duplicateConf := conf
				duplicateConf.RxMetadata = mds
				ups = append(ups, MakeJoinRequest(duplicateConf))
			}
			a.So(env.AssertHandleDeviceUplinkSuccess(ctx, func(ctx context.Context, assertEvents func(...events.Event) bool) bool {
				_, a := test.MustNewTFromContext(ctx)
				return a.So(assert(ctx, assertEvents, ups...), should.BeTrue)
			}, ups...), should.BeTrue)
		},
	}), should.BeTrue)
}

func (env TestEnvironment) AssertNsJsJoin(ctx context.Context, getPeerAssert func(ctx, reqCtx context.Context, ids ttnpb.Identifiers) bool, joinAssert func(ctx, reqCtx context.Context, msg *ttnpb.JoinRequest) bool, joinResp *ttnpb.JoinResponse, err error) bool {
	test.MustTFromContext(ctx).Helper()
	return test.RunSubtestFromContext(ctx, test.SubtestConfig{
		Name: "NsJs.HandleJoin",
		Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
			t.Helper()

			joinReqCh := make(chan NsJsHandleJoinRequest)
			if !a.So(test.AssertClusterGetPeerRequest(ctx, env.Cluster.GetPeer, func(ctx, reqCtx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) (test.ClusterGetPeerResponse, bool) {
				_, a := test.MustNewTFromContext(ctx)
				return test.ClusterGetPeerResponse{
						Peer: NewJSPeer(ctx, &MockNsJsServer{
							HandleJoinFunc: MakeNsJsHandleJoinChFunc(joinReqCh),
						}),
					}, test.AllTrue(
						a.So(role, should.Equal, ttnpb.ClusterRole_JOIN_SERVER),
						getPeerAssert(ctx, reqCtx, ids),
					)
			}), should.BeTrue) {
				return
			}
			if !a.So(test.AssertClusterAuthRequest(ctx, env.Cluster.Auth, &grpc.EmptyCallOption{}), should.BeTrue) {
				return
			}
			select {
			case <-ctx.Done():
				t.Error("Timed out while waiting for NsJs.HandleJoin to be called")
				return

			case req := <-joinReqCh:
				if !a.So(joinAssert(ctx, req.Context, req.Message), should.BeTrue) {
					return
				}
				select {
				case <-ctx.Done():
					t.Error("Timed out while waiting for NsJs.HandleJoin response to be processed")
					return

				case req.Response <- NsJsHandleJoinResponse{
					Response: joinResp,
					Error:    err,
				}:
				}
			}
		},
	})
}

type JoinAssertionConfig struct {
	Link           ttnpb.AsNs_LinkApplicationClient
	Device         *ttnpb.EndDevice
	ChannelIndex   uint8
	DataRateIndex  ttnpb.DataRateIndex
	RxMetadatas    [][]*ttnpb.RxMetadata
	CorrelationIDs []string

	ClusterResponse *NsJsHandleJoinResponse
	InteropResponse *InteropClientHandleJoinRequestResponse
}

func (env TestEnvironment) AssertJoin(ctx context.Context, conf JoinAssertionConfig) (*ttnpb.EndDevice, bool) {
	t, a := test.MustNewTFromContext(ctx)
	t.Helper()

	fp := FrequencyPlan(conf.Device.FrequencyPlanID)
	phy := LoRaWANBands[fp.BandID][conf.Device.LoRaWANPHYVersion]
	upCh := phy.UplinkChannels[conf.ChannelIndex]
	upDR := phy.DataRates[conf.DataRateIndex].Rate

	devNonce := types.DevNonce{0x42, 0x42}
	mic := [4]byte{0x42, 0x42, 0x42, 0x42}

	start := time.Now().UTC()

	upConf := JoinRequestConfig{
		JoinEUI:        *conf.Device.JoinEUI,
		DevEUI:         *conf.Device.DevEUI,
		DevNonce:       devNonce,
		DataRate:       upDR,
		Frequency:      upCh.Frequency,
		RxMetadata:     conf.RxMetadatas[0],
		CorrelationIDs: conf.CorrelationIDs,
		MIC:            mic,
	}
	var dev *ttnpb.EndDevice
	if !a.So(env.AssertHandleJoinRequest(
		ctx,
		upConf,
		func(ctx context.Context, assertEvents func(...events.Event) bool, ups ...*ttnpb.UplinkMessage) bool {
			t, a := test.MustNewTFromContext(ctx)
			t.Helper()

			defaultMACSettings := env.Config.DefaultMACSettings.Parse()

			defaultLoRaWANVersion := mac.DeviceDefaultLoRaWANVersion(conf.Device)

			defaultRX1DROffset := mac.DeviceDefaultRX1DataRateOffset(conf.Device, defaultMACSettings)
			defaultRX2DRIdx := mac.DeviceDefaultRX2DataRateIndex(conf.Device, phy, defaultMACSettings)
			defaultRX2Freq := mac.DeviceDefaultRX2Frequency(conf.Device, phy, defaultMACSettings)

			desiredRX1Delay := mac.DeviceDesiredRX1Delay(conf.Device, phy, defaultMACSettings)
			desiredRX1DROffset := mac.DeviceDesiredRX1DataRateOffset(conf.Device, defaultMACSettings)
			desiredRX2DRIdx := mac.DeviceDesiredRX2DataRateIndex(conf.Device, phy, fp, defaultMACSettings)

			deduplicatedUpConf := upConf
			deduplicatedUpConf.DecodePayload = true
			deduplicatedUpConf.ChannelIndex = conf.ChannelIndex
			deduplicatedUpConf.DataRateIndex = conf.DataRateIndex
			for _, up := range ups[1:] {
				deduplicatedUpConf.RxMetadata = append(deduplicatedUpConf.RxMetadata, up.RxMetadata...)
			}
			var joinReq *ttnpb.JoinRequest
			var joinResp *ttnpb.JoinResponse
			if conf.ClusterResponse != nil {
				if !a.So(env.AssertNsJsJoin(
					ctx,
					func(ctx, reqCtx context.Context, peerIDs ttnpb.Identifiers) bool {
						return test.AllTrue(
							a.So(events.CorrelationIDsFromContext(reqCtx), should.BeProperSupersetOfElementsFunc, test.StringEqual, ups[0].CorrelationIDs),
							a.So(peerIDs, should.Resemble, conf.Device.EndDeviceIdentifiers),
						)
					},
					func(ctx, reqCtx context.Context, req *ttnpb.JoinRequest) bool {
						joinReq = req
						return test.AllTrue(
							a.So(events.CorrelationIDsFromContext(reqCtx), should.NotBeEmpty),
							a.So(req.DevAddr, should.NotBeEmpty),
							a.So(req.DevAddr.NwkID(), should.Resemble, env.Config.NetID.ID()),
							a.So(req.DevAddr.NetIDType(), should.Equal, env.Config.NetID.Type()),
							a.So(req.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual, ups[0].CorrelationIDs),
							a.So(req, should.Resemble, MakeNsJsJoinRequest(NsJsJoinRequestConfig{
								JoinEUI:            *conf.Device.JoinEUI,
								DevEUI:             *conf.Device.DevEUI,
								DevNonce:           devNonce,
								MIC:                mic,
								DevAddr:            req.DevAddr,
								SelectedMACVersion: defaultLoRaWANVersion,
								NetID:              env.Config.NetID,
								RX1DataRateOffset:  defaultRX1DROffset,
								RX2DataRateIndex:   defaultRX2DRIdx,
								RXDelay:            desiredRX1Delay,
								FrequencyPlanID:    conf.Device.FrequencyPlanID,
								PHYVersion:         conf.Device.LoRaWANPHYVersion,
								CorrelationIDs:     req.CorrelationIDs,
							})),
						)
					},
					conf.ClusterResponse.Response,
					conf.ClusterResponse.Error,
				), should.BeTrue) {
					return false
				}
				if conf.ClusterResponse.Error == nil {
					joinResp = conf.ClusterResponse.Response
				}
			}
			if conf.InteropResponse != nil {
				t.Fatal("Interop join assertion not implemented yet")
				return false
			}

			dev = CopyEndDevice(conf.Device)
			dev.PendingMACState = &ttnpb.MACState{
				CurrentParameters: ttnpb.MACParameters{
					MaxEIRP:                    phy.DefaultMaxEIRP,
					ADRDataRateIndex:           ttnpb.DATA_RATE_0,
					ADRNbTrans:                 1,
					Rx1Delay:                   mac.DeviceDefaultRX1Delay(dev, phy, defaultMACSettings),
					Rx1DataRateOffset:          defaultRX1DROffset,
					Rx2DataRateIndex:           defaultRX2DRIdx,
					Rx2Frequency:               defaultRX2Freq,
					MaxDutyCycle:               mac.DeviceDefaultMaxDutyCycle(dev, defaultMACSettings),
					RejoinTimePeriodicity:      ttnpb.REJOIN_TIME_0,
					RejoinCountPeriodicity:     ttnpb.REJOIN_COUNT_16,
					PingSlotFrequency:          mac.DeviceDefaultPingSlotFrequency(dev, phy, defaultMACSettings),
					BeaconFrequency:            mac.DeviceDefaultBeaconFrequency(dev, defaultMACSettings),
					Channels:                   mac.DeviceDefaultChannels(dev, phy, defaultMACSettings),
					ADRAckLimitExponent:        &ttnpb.ADRAckLimitExponentValue{Value: phy.ADRAckLimit},
					ADRAckDelayExponent:        &ttnpb.ADRAckDelayExponentValue{Value: phy.ADRAckDelay},
					PingSlotDataRateIndexValue: mac.DeviceDefaultPingSlotDataRateIndexValue(dev, phy, defaultMACSettings),
				},
				DesiredParameters: ttnpb.MACParameters{
					MaxEIRP:                    mac.DeviceDesiredMaxEIRP(dev, phy, fp, defaultMACSettings),
					ADRDataRateIndex:           ttnpb.DATA_RATE_0,
					ADRNbTrans:                 1,
					Rx1Delay:                   desiredRX1Delay,
					Rx1DataRateOffset:          desiredRX1DROffset,
					Rx2DataRateIndex:           desiredRX2DRIdx,
					Rx2Frequency:               mac.DeviceDesiredRX2Frequency(dev, phy, fp, defaultMACSettings),
					MaxDutyCycle:               mac.DeviceDesiredMaxDutyCycle(dev, defaultMACSettings),
					RejoinTimePeriodicity:      ttnpb.REJOIN_TIME_0,
					RejoinCountPeriodicity:     ttnpb.REJOIN_COUNT_16,
					PingSlotFrequency:          mac.DeviceDesiredPingSlotFrequency(dev, phy, fp, defaultMACSettings),
					BeaconFrequency:            mac.DeviceDesiredBeaconFrequency(dev, defaultMACSettings),
					Channels:                   mac.DeviceDesiredChannels(phy, fp, defaultMACSettings),
					UplinkDwellTime:            mac.DeviceDesiredUplinkDwellTime(fp),
					DownlinkDwellTime:          mac.DeviceDesiredDownlinkDwellTime(fp),
					ADRAckLimitExponent:        mac.DeviceDesiredADRAckLimitExponent(dev, phy, defaultMACSettings),
					ADRAckDelayExponent:        mac.DeviceDesiredADRAckDelayExponent(dev, phy, defaultMACSettings),
					PingSlotDataRateIndexValue: mac.DeviceDesiredPingSlotDataRateIndexValue(dev, phy, fp, defaultMACSettings),
				},
				DeviceClass:    test.Must(mac.DeviceDefaultClass(dev)).(ttnpb.Class),
				LoRaWANVersion: defaultLoRaWANVersion,
				QueuedJoinAccept: &ttnpb.MACState_JoinAccept{
					Payload: joinResp.RawPayload,
					Request: *joinReq,
					Keys: func() ttnpb.SessionKeys {
						keys := ttnpb.SessionKeys{
							SessionKeyID: joinResp.SessionKeys.SessionKeyID,
							FNwkSIntKey:  joinResp.SessionKeys.FNwkSIntKey,
							NwkSEncKey:   joinResp.SessionKeys.NwkSEncKey,
							SNwkSIntKey:  joinResp.SessionKeys.SNwkSIntKey,
						}
						if !joinReq.DownlinkSettings.OptNeg {
							keys.NwkSEncKey = keys.FNwkSIntKey
							keys.SNwkSIntKey = keys.FNwkSIntKey
						}
						return *CopySessionKeys(&keys)
					}(),
					CorrelationIDs: joinResp.CorrelationIDs,
				},
				RxWindowsAvailable: true,
			}
			dev.RecentUplinks = AppendRecentUplink(dev.RecentUplinks, MakeJoinRequest(deduplicatedUpConf), RecentUplinkCount)

			idsWithDevAddr := conf.Device.EndDeviceIdentifiers
			idsWithDevAddr.DevAddr = &joinReq.DevAddr

			if !a.So(assertEvents(events.Builders(func() []events.Builder {
				evBuilders := []events.Builder{
					EvtReceiveJoinRequest,
				}
				for range ups[1:] {
					evBuilders = append(evBuilders,
						EvtReceiveJoinRequest,
						EvtDropJoinRequest.With(events.WithData(ErrDuplicate)),
					)
				}
				if conf.ClusterResponse != nil {
					evBuilders = append(evBuilders,
						EvtClusterJoinAttempt,
					)
					if conf.ClusterResponse.Error == nil {
						evBuilders = append(evBuilders,
							EvtClusterJoinSuccess.With(events.WithData(JoinResponseWithoutKeys(conf.ClusterResponse.Response))),
						)
					}
				}
				return append(evBuilders,
					EvtProcessJoinRequest,
				)
			}()).New(
				ctx,
				events.WithIdentifiers(conf.Device.EndDeviceIdentifiers),
			)...), should.BeTrue) {
				return false
			}

			var appUp *ttnpb.ApplicationUp
			if !a.So(AssertProcessApplicationUp(ctx, conf.Link, func(ctx context.Context, up *ttnpb.ApplicationUp) bool {
				_, a := test.MustNewTFromContext(ctx)
				recvAt := up.GetJoinAccept().GetReceivedAt()
				appUp = up
				return test.AllTrue(
					a.So(up.CorrelationIDs, should.HaveSameElementsDeep, append(joinReq.CorrelationIDs, joinResp.CorrelationIDs...)),
					a.So([]time.Time{start, recvAt, time.Now()}, should.BeChronological),
					a.So(up, should.Resemble, &ttnpb.ApplicationUp{
						EndDeviceIdentifiers: idsWithDevAddr,
						CorrelationIDs:       up.CorrelationIDs,
						Up: &ttnpb.ApplicationUp_JoinAccept{
							JoinAccept: &ttnpb.ApplicationJoinAccept{
								AppSKey:      joinResp.AppSKey,
								SessionKeyID: joinResp.SessionKeyID,
								ReceivedAt:   recvAt,
							},
						},
					}),
				)
			}), should.BeTrue) {
				t.Error("Failed to send join-accept to Application Server")
				return false
			}
			return a.So(env.Events, should.ReceiveEventFunc, test.MakeEventEqual(test.EventEqualConfig{
				Identifiers:    true,
				Data:           true,
				Origin:         true,
				Context:        true,
				Visibility:     true,
				Authentication: true,
				RemoteIP:       true,
				UserAgent:      true,
			}),
				EvtForwardJoinAccept.NewWithIdentifiersAndData(conf.Link.Context(), idsWithDevAddr, &ttnpb.ApplicationUp{
					EndDeviceIdentifiers: idsWithDevAddr,
					CorrelationIDs:       appUp.CorrelationIDs,
					Up: &ttnpb.ApplicationUp_JoinAccept{
						JoinAccept: ApplicationJoinAcceptWithoutAppSKey(appUp.GetJoinAccept()),
					},
				}),
			)
		},
		conf.RxMetadatas[1:]...,
	), should.BeTrue) {
		return nil, false
	}
	return env.AssertScheduleJoinAccept(ctx, dev)
}

type DataUplinkAssertionConfig struct {
	Link           ttnpb.AsNs_LinkApplicationClient
	Device         *ttnpb.EndDevice
	ChannelIndex   uint8
	DataRateIndex  ttnpb.DataRateIndex
	RxMetadatas    [][]*ttnpb.RxMetadata
	CorrelationIDs []string

	Confirmed    bool
	Pending      bool
	FRMPayload   []byte
	FOpts        []byte
	FCtrl        ttnpb.FCtrl
	FCntDelta    uint32
	ConfFCntDown uint32
	FPort        uint8

	EventBuilders []events.Builder
}

func (env TestEnvironment) AssertHandleDataUplink(ctx context.Context, conf DataUplinkAssertionConfig) (*ttnpb.EndDevice, bool) {
	t, a := test.MustNewTFromContext(ctx)
	t.Helper()

	dev := CopyEndDevice(conf.Device)
	return dev, a.So(test.RunSubtestFromContext(ctx, test.SubtestConfig{
		Name: "Data uplink",
		Func: func(ctx context.Context, t *testing.T, a *assertions.Assertion) {
			t.Helper()

			start := time.Now().UTC()
			upConf := WithDeviceDataUplinkConfig(dev, conf.Pending, conf.DataRateIndex, conf.ChannelIndex, conf.FCntDelta)(DataUplinkConfig{
				Confirmed:      conf.Confirmed,
				FCtrl:          conf.FCtrl,
				ConfFCntDown:   conf.ConfFCntDown,
				FPort:          conf.FPort,
				FRMPayload:     conf.FRMPayload,
				FOpts:          conf.FOpts,
				RxMetadata:     conf.RxMetadatas[0],
				CorrelationIDs: conf.CorrelationIDs,
			})

			deduplicatedUpConf := upConf
			deduplicatedUpConf.DecodePayload = true
			deduplicatedUpConf.Matched = true
			ups := []*ttnpb.UplinkMessage{MakeDataUplink(upConf)}
			for _, mds := range conf.RxMetadatas[1:] {
				mds := mds
				duplicateConf := upConf
				duplicateConf.RxMetadata = mds
				ups = append(ups, MakeDataUplink(duplicateConf))
				deduplicatedUpConf.RxMetadata = append(deduplicatedUpConf.RxMetadata, mds...)
			}
			if !a.So(env.AssertHandleDeviceUplinkSuccess(ctx, func(ctx context.Context, assertEvents func(...events.Event) bool) bool {
				t, a := test.MustNewTFromContext(ctx)
				t.Helper()
				if !a.So(assertEvents(events.Builders(func() []events.Builder {
					evBuilders := []events.Builder{
						EvtReceiveDataUplink,
					}
					for range ups[1:] {
						evBuilders = append(evBuilders,
							EvtReceiveDataUplink,
							EvtDropDataUplink.With(events.WithData(ErrDuplicate)),
						)
					}
					return append(
						append(
							evBuilders,
							conf.EventBuilders...,
						),
						EvtProcessDataUplink,
					)
				}()).New(
					ctx,
					events.WithIdentifiers(dev.EndDeviceIdentifiers),
				)...), should.BeTrue) {
					t.Error("Uplink event assertion failed")
					return false
				}
				return true
			}, ups...), should.BeTrue) {
				t.Error("Data uplink send assertion failed")
				return
			}

			deduplicatedUp := MakeDataUplink(deduplicatedUpConf)
			var appUp *ttnpb.ApplicationUp
			if !a.So(AssertProcessApplicationUp(ctx, conf.Link, func(ctx context.Context, up *ttnpb.ApplicationUp) bool {
				_, a := test.MustNewTFromContext(ctx)
				recvAt := up.GetUplinkMessage().GetReceivedAt()
				appUp = up
				return test.AllTrue(
					a.So(up.CorrelationIDs, should.BeProperSupersetOfElementsFunc, test.StringEqual, deduplicatedUp.CorrelationIDs),
					a.So(up.GetUplinkMessage().GetRxMetadata(), should.HaveSameElementsDeep, deduplicatedUp.RxMetadata),
					a.So([]time.Time{start, recvAt, time.Now()}, should.BeChronological),
					a.So(up, should.Resemble, &ttnpb.ApplicationUp{
						EndDeviceIdentifiers: dev.EndDeviceIdentifiers,
						CorrelationIDs:       up.CorrelationIDs,
						Up: &ttnpb.ApplicationUp_UplinkMessage{
							UplinkMessage: &ttnpb.ApplicationUplink{
								Confirmed:    conf.Confirmed,
								FPort:        deduplicatedUp.Payload.GetMACPayload().FPort,
								FRMPayload:   deduplicatedUp.Payload.GetMACPayload().FRMPayload,
								ReceivedAt:   up.GetUplinkMessage().GetReceivedAt(),
								RxMetadata:   up.GetUplinkMessage().GetRxMetadata(),
								Settings:     deduplicatedUp.Settings,
								SessionKeyID: upConf.SessionKeys.SessionKeyID,
							},
						},
					}),
				)
			}), should.BeTrue) {
				t.Error("Application Server data uplink forwarding assertion failed")
				return
			}
			if !a.So(env.Events, should.ReceiveEventFunc, test.MakeEventEqual(test.EventEqualConfig{
				Identifiers:    true,
				Data:           true,
				Origin:         true,
				Context:        true,
				Visibility:     true,
				Authentication: true,
				RemoteIP:       true,
				UserAgent:      true,
			}),
				EvtForwardDataUplink.New(
					conf.Link.Context(),
					events.WithIdentifiers(dev.EndDeviceIdentifiers),
					events.WithData(appUp),
				),
			) {
				t.Error("Application Server forwarding event assertion failed")
			}
			if conf.Pending {
				dev.MACState = dev.PendingMACState
				dev.MACState.CurrentParameters.Rx1Delay = dev.PendingMACState.PendingJoinRequest.RxDelay
				dev.MACState.CurrentParameters.Rx1DataRateOffset = dev.PendingMACState.PendingJoinRequest.DownlinkSettings.Rx1DROffset
				dev.MACState.CurrentParameters.Rx2DataRateIndex = dev.PendingMACState.PendingJoinRequest.DownlinkSettings.Rx2DR
				dev.MACState.PendingJoinRequest = nil
				dev.Session = dev.PendingSession
				dev.PendingMACState = nil
				dev.PendingSession = nil
			}
			dev.RecentUplinks = AppendRecentUplink(dev.RecentUplinks, deduplicatedUp, RecentUplinkCount)
			dev.MACState.RecentUplinks = AppendRecentUplink(dev.MACState.RecentUplinks, deduplicatedUp, RecentUplinkCount)
		},
	}), should.BeTrue)
}

func DownlinkProtoPaths(paths ...DownlinkPath) (pbs []*ttnpb.DownlinkPath) {
	for _, p := range paths {
		pbs = append(pbs, p.DownlinkPath)
	}
	return pbs
}

func (env TestEnvironment) AssertSetDevice(ctx context.Context, create bool, req *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, bool) {
	t, a := test.MustNewTFromContext(ctx)
	t.Helper()

	listRightsCh := make(chan test.ApplicationAccessListRightsRequest)
	defer func() {
		close(listRightsCh)
	}()

	var dev *ttnpb.EndDevice
	var err error
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		dev, err = ttnpb.NewNsEndDeviceRegistryClient(env.ClientConn).Set(
			ctx,
			req,
			grpc.PerRPCCredentials(rpcmetadata.MD{
				AuthType:      "Bearer",
				AuthValue:     "set-key",
				AllowInsecure: true,
			}),
		)
		wg.Done()
	}()

	var reqCIDs []string
	if !a.So(test.AssertClusterGetPeerRequest(ctx, env.Cluster.GetPeer,
		func(ctx, reqCtx context.Context, role ttnpb.ClusterRole, ids ttnpb.Identifiers) (test.ClusterGetPeerResponse, bool) {
			_, a := test.MustNewTFromContext(ctx)
			reqCIDs = events.CorrelationIDsFromContext(reqCtx)
			return test.ClusterGetPeerResponse{
					Peer: NewISPeer(ctx, &test.MockApplicationAccessServer{
						ListRightsFunc: test.MakeApplicationAccessListRightsChFunc(listRightsCh),
					}),
				}, test.AllTrue(
					a.So(role, should.Equal, ttnpb.ClusterRole_ACCESS),
					a.So(ids, should.BeNil),
				)
		}), should.BeTrue) {
		return nil, false
	}
	a.So(reqCIDs, should.HaveLength, 1)

	if !a.So(test.AssertListRightsRequest(ctx, listRightsCh,
		func(ctx, reqCtx context.Context, ids ttnpb.Identifiers) bool {
			_, a := test.MustNewTFromContext(ctx)
			md := rpcmetadata.FromIncomingContext(reqCtx)
			return a.So(md.AuthType, should.Equal, "Bearer") &&
				a.So(md.AuthValue, should.Equal, "set-key") &&
				a.So(ids, should.Resemble, &req.EndDevice.ApplicationIdentifiers)
		}, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
	), should.BeTrue) {
		return nil, false
	}

	ev := EvtCreateEndDevice.BindData(nil)
	if !create {
		ev = EvtUpdateEndDevice.BindData(nil)
	}
	if !a.So(env.Events, should.ReceiveEventResembling, ev.New(events.ContextWithCorrelationID(test.Context(), reqCIDs...), events.WithIdentifiers(req.EndDevice.EndDeviceIdentifiers))) {
		if create {
			t.Error("Failed to assert end device create event")
		} else {
			t.Error("Failed to assert end device update event")
		}
		return nil, false
	}

	if !a.So(test.WaitContext(ctx, wg.Wait), should.BeTrue) {
		t.Error("Timed out while waiting for device to be set")
		return nil, false
	}
	return dev, a.So(err, should.BeNil)
}

func StartTaskExclude(names ...string) component.StartTaskFunc {
	if !sort.StringsAreSorted(names) {
		panic("names must be sorted alphabetically")
	}
	return func(conf *component.TaskConfig) {
		for _, name := range names {
			if i := sort.Search(len(names), func(i int) bool {
				return names[i] == name
			}); i < len(names) && names[i] == name {
				return
			}
		}
		component.DefaultStartTask(conf)
	}
}

type TestConfig struct {
	Context              context.Context
	NetworkServer        Config
	NetworkServerOptions []Option
	Component            component.Config
	TaskStarter          component.TaskStarter
}

func StartTest(t *testing.T, conf TestConfig) (*NetworkServer, context.Context, TestEnvironment, func()) {
	t.Helper()

	authCh := make(chan test.ClusterAuthRequest)
	getPeerCh := make(chan test.ClusterGetPeerRequest)
	eventsPublishCh := make(chan test.EventPubSubPublishRequest)

	var closeFuncs []func()
	closeFuncs = append(closeFuncs, test.SetDefaultEventsPubSub(&test.MockEventPubSub{
		PublishFunc: test.MakeEventPubSubPublishChFunc(eventsPublishCh),
	}))
	if conf.NetworkServer.DeduplicationWindow == 0 {
		conf.NetworkServer.DeduplicationWindow = time.Nanosecond
	}
	if conf.NetworkServer.CooldownWindow == 0 {
		conf.NetworkServer.CooldownWindow = conf.NetworkServer.DeduplicationWindow + time.Nanosecond
	}

	cmpOpts := []component.Option{
		component.WithClusterNew(func(context.Context, *cluster.Config, ...cluster.Option) (cluster.Cluster, error) {
			return &test.MockCluster{
				AuthFunc:    test.MakeClusterAuthChFunc(authCh),
				GetPeerFunc: test.MakeClusterGetPeerChFunc(getPeerCh),
				JoinFunc:    test.ClusterJoinNilFunc,
				WithVerifiedSourceFunc: func(ctx context.Context) context.Context {
					return clusterauth.NewContext(ctx, nil)
				},
			}, nil
		}),
	}
	if conf.TaskStarter != nil {
		cmpOpts = append(cmpOpts, component.WithTaskStarter(conf.TaskStarter))
	}

	if conf.NetworkServer.Devices == nil {
		v, closeFn := NewDeviceRegistry(t)
		if closeFn != nil {
			closeFuncs = append(closeFuncs, closeFn)
		}
		conf.NetworkServer.Devices = v
	}
	if conf.NetworkServer.ApplicationUplinkQueue.Queue == nil {
		v, closeFn := NewApplicationUplinkQueue(t)
		if closeFn != nil {
			closeFuncs = append(closeFuncs, closeFn)
		}
		conf.NetworkServer.ApplicationUplinkQueue.Queue = v
	}
	if conf.NetworkServer.DownlinkTasks == nil {
		v, closeFn := NewDownlinkTaskQueue(t)
		if closeFn != nil {
			closeFuncs = append(closeFuncs, closeFn)
		}
		conf.NetworkServer.DownlinkTasks = v
	}
	if conf.NetworkServer.UplinkDeduplicator == nil {
		v, closeFn := NewUplinkDeduplicator(t)
		if closeFn != nil {
			closeFuncs = append(closeFuncs, closeFn)
		}
		conf.NetworkServer.UplinkDeduplicator = v
	}

	ns := test.Must(New(
		componenttest.NewComponent(t, &conf.Component, cmpOpts...),
		&conf.NetworkServer,
		conf.NetworkServerOptions...,
	)).(*NetworkServer)
	ns.Component.FrequencyPlans = frequencyplans.NewStore(test.FrequencyPlansFetcher)

	env := TestEnvironment{
		Config: conf.NetworkServer,

		Cluster: TestClusterEnvironment{
			Auth:    authCh,
			GetPeer: getPeerCh,
		},
		Events: eventsPublishCh,
	}
	if ns.interopClient == nil {
		m, mEnv, closeM := newMockInteropClient(t)
		ns.interopClient = m
		env.InteropClient = &mEnv
		closeFuncs = append(closeFuncs, closeM)
	}

	componenttest.StartComponent(t, ns.Component)
	env.ClientConn = ns.LoopbackConn()

	ctx, cancel := context.WithCancel(conf.Context)
	return ns, ctx, env, func() {
		cancel()
		for _, f := range closeFuncs {
			f()
		}
		select {
		case <-authCh:
			t.Error("Cluster.Auth call missed")
		default:
			close(authCh)
		}
		select {
		case <-getPeerCh:
			t.Error("Cluster.GetPeer call missed")
		default:
			close(getPeerCh)
		}
		select {
		case <-eventsPublishCh:
			t.Error("events.Publish call missed")
		default:
			close(eventsPublishCh)
		}
	}
}

func LogEvents(t *testing.T, ch <-chan test.EventPubSubPublishRequest) {
	for ev := range ch {
		t.Logf("Event %s published with data %v", ev.Event.Name(), ev.Event.Data())
		ev.Response <- struct{}{}
	}
}

func MustCreateDevice(ctx context.Context, r DeviceRegistry, dev *ttnpb.EndDevice, paths ...string) (*ttnpb.EndDevice, context.Context) {
	dev, ctx, err := CreateDevice(ctx, r, dev, paths...)
	test.Must(nil, err)
	return dev, ctx
}

func MustGetDeviceByID(ctx context.Context, r DeviceRegistry, appID ttnpb.ApplicationIdentifiers, devID string, paths ...string) (*ttnpb.EndDevice, context.Context) {
	if len(paths) == 0 {
		paths = ttnpb.EndDeviceFieldPathsTopLevel
	}
	dev, ctx, err := r.GetByID(ctx, appID, devID, paths)
	test.Must(nil, err)
	return dev, ctx
}

type SetDeviceRequest struct {
	*ttnpb.EndDevice
	Paths []string
}

type ContextualEndDevice struct {
	context.Context
	*ttnpb.EndDevice
}

func MustCreateDevices(ctx context.Context, r DeviceRegistry, devs ...SetDeviceRequest) []*ContextualEndDevice {
	var setDevices []*ContextualEndDevice
	for _, dev := range devs {
		set, ctx := MustCreateDevice(ctx, r, dev.EndDevice, dev.Paths...)
		setDevices = append(setDevices, &ContextualEndDevice{
			Context:   ctx,
			EndDevice: set,
		})
	}
	return setDevices
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

var _ DeviceRegistry = MockDeviceRegistry{}

// MockDeviceRegistry is a mock DeviceRegistry used for testing.
type MockDeviceRegistry struct {
	GetByIDFunc     func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, context.Context, error)
	SetByIDFunc     func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error)
	RangeByAddrFunc func(ctx context.Context, devAddr types.DevAddr, paths []string, f func(context.Context, *ttnpb.EndDevice) bool) error
}

// GetByEUI panics.
func (m MockDeviceRegistry) GetByEUI(ctx context.Context, joinEUI, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, context.Context, error) {
	panic("GetByEUI must not be called")
}

// GetByID calls GetByIDFunc if set and panics otherwise.
func (m MockDeviceRegistry) GetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, context.Context, error) {
	if m.GetByIDFunc == nil {
		panic("GetByID called, but not set")
	}
	return m.GetByIDFunc(ctx, appID, devID, paths)
}

// RangeByAddr calls RangeByAddrFunc if set and panics otherwise.
func (m MockDeviceRegistry) RangeByAddr(ctx context.Context, devAddr types.DevAddr, paths []string, f func(context.Context, *ttnpb.EndDevice) bool) error {
	if m.RangeByAddrFunc == nil {
		panic("RangeByAddr called, but not set")
	}
	return m.RangeByAddrFunc(ctx, devAddr, paths, f)
}

// SetByID calls SetByIDFunc if set and panics otherwise.
func (m MockDeviceRegistry) SetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(context.Context, *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, context.Context, error) {
	if m.SetByIDFunc == nil {
		panic("SetByID called, but not set")
	}
	return m.SetByIDFunc(ctx, appID, devID, paths, f)
}

var _ UplinkDeduplicator = &MockUplinkDeduplicator{}

type MockUplinkDeduplicator struct {
	DeduplicateUplinkFunc   func(context.Context, *ttnpb.UplinkMessage, time.Duration) (bool, error)
	AccumulatedMetadataFunc func(context.Context, *ttnpb.UplinkMessage) ([]*ttnpb.RxMetadata, error)
}

// DeduplicateUplink calls DeduplicateUplinkFunc if set and panics otherwise.
func (m MockUplinkDeduplicator) DeduplicateUplink(ctx context.Context, up *ttnpb.UplinkMessage, d time.Duration) (bool, error) {
	if m.DeduplicateUplinkFunc == nil {
		panic("DeduplicateUplink called, but not set")
	}
	return m.DeduplicateUplinkFunc(ctx, up, d)
}

// AccumulatedMetadata calls AccumulatedMetadataFunc if set and panics otherwise.
func (m MockUplinkDeduplicator) AccumulatedMetadata(ctx context.Context, up *ttnpb.UplinkMessage) ([]*ttnpb.RxMetadata, error) {
	if m.AccumulatedMetadataFunc == nil {
		panic("AccumulatedMetadata called, but not set")
	}
	return m.AccumulatedMetadataFunc(ctx, up)
}

type UplinkDeduplicatorDeduplicateUplinkResponse struct {
	Ok    bool
	Error error
}

type UplinkDeduplicatorDeduplicateUplinkRequest struct {
	Context  context.Context
	Uplink   *ttnpb.UplinkMessage
	Window   time.Duration
	Response chan<- UplinkDeduplicatorDeduplicateUplinkResponse
}

func MakeUplinkDeduplicatorDeduplicateUplinkChFunc(reqCh chan<- UplinkDeduplicatorDeduplicateUplinkRequest) func(context.Context, *ttnpb.UplinkMessage, time.Duration) (bool, error) {
	return func(ctx context.Context, up *ttnpb.UplinkMessage, window time.Duration) (bool, error) {
		respCh := make(chan UplinkDeduplicatorDeduplicateUplinkResponse)
		reqCh <- UplinkDeduplicatorDeduplicateUplinkRequest{
			Context:  ctx,
			Uplink:   up,
			Window:   window,
			Response: respCh,
		}
		resp := <-respCh
		return resp.Ok, resp.Error
	}
}

type UplinkDeduplicatorAccumulatedMetadataResponse struct {
	Metadata []*ttnpb.RxMetadata
	Error    error
}

type UplinkDeduplicatorAccumulatedMetadataRequest struct {
	Context  context.Context
	Uplink   *ttnpb.UplinkMessage
	Response chan<- UplinkDeduplicatorAccumulatedMetadataResponse
}

func MakeUplinkDeduplicatorAccumulatedMetadataChFunc(reqCh chan<- UplinkDeduplicatorAccumulatedMetadataRequest) func(context.Context, *ttnpb.UplinkMessage) ([]*ttnpb.RxMetadata, error) {
	return func(ctx context.Context, up *ttnpb.UplinkMessage) ([]*ttnpb.RxMetadata, error) {
		respCh := make(chan UplinkDeduplicatorAccumulatedMetadataResponse)
		reqCh <- UplinkDeduplicatorAccumulatedMetadataRequest{
			Context:  ctx,
			Uplink:   up,
			Response: respCh,
		}
		resp := <-respCh
		return resp.Metadata, resp.Error
	}
}

func AssertDeduplicateUplink(ctx context.Context, reqCh <-chan UplinkDeduplicatorDeduplicateUplinkRequest, assert func(context.Context, *ttnpb.UplinkMessage, time.Duration) bool, resp UplinkDeduplicatorDeduplicateUplinkResponse) bool {
	t := test.MustTFromContext(ctx)
	t.Helper()
	select {
	case <-ctx.Done():
		t.Error("Timed out while waiting for UplinkDeduplicator.DeduplicateUplink to be called")
		return false

	case req := <-reqCh:
		if !assert(req.Context, req.Uplink, req.Window) {
			return false
		}
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for UplinkDeduplicator.DeduplicateUplink response to be processed")
			return false

		case req.Response <- resp:
			return true
		}
	}
}

func AssertAccumulatedMetadata(ctx context.Context, reqCh <-chan UplinkDeduplicatorAccumulatedMetadataRequest, assert func(context.Context, *ttnpb.UplinkMessage) bool, resp UplinkDeduplicatorAccumulatedMetadataResponse) bool {
	t := test.MustTFromContext(ctx)
	t.Helper()
	select {
	case <-ctx.Done():
		t.Error("Timed out while waiting for UplinkDeduplicator.AccumulatedMetadata to be called")
		return false

	case req := <-reqCh:
		if !assert(req.Context, req.Uplink) {
			return false
		}
		select {
		case <-ctx.Done():
			t.Error("Timed out while waiting for UplinkDeduplicator.AccumulatedMetadata response to be processed")
			return false

		case req.Response <- resp:
			return true
		}
	}
}
