// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package lbslns

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/gatewayserver/io/ws"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
)

type lbsLNS struct {
	tokens   io.DownlinkTokens
	sessions ws.Sessions
}

// NewFormatter returns a new LoRa Basic Station LNS formatter.
func NewFormatter() ws.Formatter {
	return &lbsLNS{}
}

func (f *lbsLNS) Connect(ctx context.Context, uid string) error {
	return f.sessions.NewSession(ctx, uid)
}

func (f *lbsLNS) Disconnect(ctx context.Context, uid string) {
	err := f.sessions.DeleteSession(uid)
	if err != nil {
		logger := log.FromContext(ctx)
		logger.WithError(err).Warn("Failed to disconnect")
	}
}
