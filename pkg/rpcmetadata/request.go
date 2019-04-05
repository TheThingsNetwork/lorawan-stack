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

package rpcmetadata

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"google.golang.org/grpc"
)

// GetRequestMetadata returns the request metadata with per-rpc credentials
func (m MD) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	md := make(map[string]string)
	if m.ID != "" {
		md["id"] = m.ID
	}
	if m.AuthType != "" && m.AuthValue != "" {
		md["authorization"] = m.AuthType + " " + m.AuthValue
	}
	return md, nil
}

var errUnauthenticated = errors.DefineUnauthenticated("unauthenticated", "the context is not authenticated")

// WithForwardedAuth returns a grpc.CallOption with authentication from the incoming context ctx.
func WithForwardedAuth(ctx context.Context, allowInsecure bool) (grpc.CallOption, error) {
	md := FromIncomingContext(ctx)
	if md.AuthType == "" || md.AuthValue == "" {
		return nil, errUnauthenticated
	}
	md.AllowInsecure = allowInsecure
	return grpc.PerRPCCredentials(md), nil
}
