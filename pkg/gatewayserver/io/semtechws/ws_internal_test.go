// Copyright © 2026 The Things Network Foundation, The Things Industries B.V.
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

package semtechws

import (
	"fmt"
	"io"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestDisconnectError(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		Name string
		Err  error
		// Code is the expected close code, nil if the error is expected to pass through.
		Code any
	}{
		{
			// This is what the gateway disappearing without a close handshake looks like.
			Name: "AbnormalClosure",
			Err:  &websocket.CloseError{Code: websocket.CloseAbnormalClosure, Text: io.ErrUnexpectedEOF.Error()},
			Code: websocket.CloseAbnormalClosure,
		},
		{
			Name: "GoingAway",
			Err:  &websocket.CloseError{Code: websocket.CloseGoingAway},
			Code: websocket.CloseGoingAway,
		},
		{
			Name: "WrappedCloseError",
			Err: fmt.Errorf("read: %w",
				&websocket.CloseError{Code: websocket.CloseNoStatusReceived},
			),
			Code: websocket.CloseNoStatusReceived,
		},
		{
			Name: "DefinedError",
			Err:  errMissedTooManyPongs.New(),
		},
		{
			Name: "OtherError",
			Err:  io.ErrUnexpectedEOF,
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			a := assertions.New(t)

			err := disconnectError(tc.Err)
			if tc.Code == nil {
				a.So(err, should.Equal, tc.Err)
				return
			}
			if !a.So(errors.IsAborted(err), should.BeTrue) {
				t.FailNow()
			}
			ttnErr, ok := errors.From(err)
			if !a.So(ok, should.BeTrue) {
				t.FailNow()
			}
			// The error must be classifiable, so that it can be used as a metric label.
			a.So(ttnErr.FullName(), should.Equal, "pkg/gatewayserver/io/semtechws:websocket_closed")
			a.So(ttnErr.Attributes()["code"], should.Equal, tc.Code)
			// The original error must be preserved for diagnostics.
			a.So(errors.Is(err, tc.Err), should.BeTrue)
		})
	}
}
