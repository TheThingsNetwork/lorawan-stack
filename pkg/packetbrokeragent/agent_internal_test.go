// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package packetbrokeragent

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc"
)

type testAuthenticator struct {
	id          *ttnpb.PacketBrokerNetworkIdentifier
	dialOptions []grpc.DialOption
}

func (a *testAuthenticator) AuthInfo(context.Context) (*ttnpb.PacketBrokerNetworkIdentifier, error) {
	return a.id, nil
}

func (a *testAuthenticator) DialOptions(context.Context) ([]grpc.DialOption, error) {
	return a.dialOptions, nil
}

func WithTestAuthenticator(id *ttnpb.PacketBrokerNetworkIdentifier) Option {
	return func(a *Agent) {
		a.authenticator = &testAuthenticator{
			id: id,
			dialOptions: []grpc.DialOption{
				grpc.WithInsecure(),
			},
		}
	}
}
