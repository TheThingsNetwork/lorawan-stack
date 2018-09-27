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

	"github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (as *ApplicationServer) GetLink(ctx context.Context, req *ttnpb.GetApplicationLinkRequest) (*ttnpb.ApplicationLink, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (as *ApplicationServer) SetLink(ctx context.Context, req *ttnpb.SetApplicationLinkRequest) (*ttnpb.ApplicationLink, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}

func (as *ApplicationServer) DeleteLink(ctx context.Context, ids *ttnpb.ApplicationIdentifiers) (*types.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "not implemented")
}
