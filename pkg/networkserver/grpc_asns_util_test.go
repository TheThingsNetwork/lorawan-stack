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
	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

var _ ttnpb.AsNs_LinkApplicationServer = &MockAsNsLinkApplicationStream{}

type MockAsNsLinkApplicationStream struct {
	*test.MockServerStream
	SendFunc func(*ttnpb.ApplicationUp) error
	RecvFunc func() (*pbtypes.Empty, error)
}

func (s *MockAsNsLinkApplicationStream) Send(msg *ttnpb.ApplicationUp) error {
	if s.SendFunc == nil {
		return nil
	}
	return s.SendFunc(msg)
}

func (s *MockAsNsLinkApplicationStream) Recv() (*pbtypes.Empty, error) {
	if s.RecvFunc == nil {
		return ttnpb.Empty, nil
	}
	return s.RecvFunc()
}
