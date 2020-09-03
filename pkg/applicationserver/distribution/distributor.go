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

package distribution

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// Distributor sends upstream messages to a set of subscriptions
// in a broadcast manner.
type Distributor interface {
	// Subscribe adds the subscription to the appropriate set.
	// The subscription is automatically removed when cancelled.
	Subscribe(context.Context, *io.Subscription) error
	// SendUp broadcasts the upstream message to the appropriate set.
	SendUp(context.Context, *ttnpb.ApplicationUp) error
}
