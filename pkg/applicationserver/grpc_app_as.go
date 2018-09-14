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

package applicationserver

import (
	"context"

	ptypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (as *ApplicationServer) Subscribe(req *ttnpb.ApplicationIdentifiers, stream ttnpb.AppAs_SubscribeServer) error {
	return status.Errorf(codes.Unimplemented, "not implemented")
}

// DownlinkQueueReplace is called by the Application Server to completely replace the downlink queue for a device.
func (as *ApplicationServer) DownlinkQueueReplace(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*ptypes.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// DownlinkQueuePush is called by the Application Server to push a downlink to queue for a device.
func (as *ApplicationServer) DownlinkQueuePush(ctx context.Context, req *ttnpb.DownlinkQueueRequest) (*ptypes.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// DownlinkQueueList is called by the Application Server to get the current state of the downlink queue for a device.
func (as *ApplicationServer) DownlinkQueueList(ctx context.Context, req *ttnpb.EndDeviceIdentifiers) (*ttnpb.ApplicationDownlinks, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

// DownlinkQueueClear is called by the Application Server to clear the downlink queue for a device.
func (as *ApplicationServer) DownlinkQueueClear(ctx context.Context, id *ttnpb.EndDeviceIdentifiers) (*ptypes.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}
