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

package joinserver

import (
	"context"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type jsEndDeviceRegistryServer struct {
	JS *JoinServer
}

func (jsEndDeviceRegistryServer) Get(context.Context, *ttnpb.GetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (jsEndDeviceRegistryServer) Set(context.Context, *ttnpb.SetEndDeviceRequest) (*ttnpb.EndDevice, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (jsEndDeviceRegistryServer) Delete(context.Context, *ttnpb.EndDeviceIdentifiers) (*pbtypes.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}
