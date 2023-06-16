// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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

package deviceclaimingserver_test

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
)

// MockClaimer is a mock Claimer.
type MockClaimer struct {
	JoinEUI types.EUI64
}

// SupportsJoinEUI returns whether the Join Server supports this JoinEUI.
func (m MockClaimer) SupportsJoinEUI(joinEUI types.EUI64) bool {
	return m.JoinEUI.Equal(joinEUI)
}

// Claim claims an End Device.
func (MockClaimer) Claim(_ context.Context, _, _ types.EUI64, _ string,
) error {
	return nil
}

// GetClaimStatus returns the claim status an End Device.
func (MockClaimer) GetClaimStatus(_ context.Context,
	ids *ttnpb.EndDeviceIdentifiers,
) (*ttnpb.GetClaimStatusResponse, error) {
	return &ttnpb.GetClaimStatusResponse{
		EndDeviceIds: ids,
	}, nil
}

// Unclaim releases the claim on an End Device.
func (MockClaimer) Unclaim(_ context.Context,
	_ *ttnpb.EndDeviceIdentifiers,
) (err error) {
	return nil
}
