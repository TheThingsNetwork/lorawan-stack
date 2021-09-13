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

package workerpool

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// HandlerFactoryFromUplinkHandler converts a static uplink handler to a HandlerFactory.
func HandlerFactoryFromUplinkHandler(handler func(context.Context, *ttnpb.ApplicationUp) error) HandlerFactory {
	h := func(ctx context.Context, item interface{}) {
		up := item.(*ttnpb.ApplicationUp)

		if err := handler(ctx, up); err != nil {
			log.FromContext(ctx).WithError(err).Warn("Failed to submit message")
		}
	}
	return StaticHandlerFactory(h)
}
