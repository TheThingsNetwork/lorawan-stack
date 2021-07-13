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

package pubsub

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmiddleware/warning"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// ProviderStatus is the status of a PubSub provider.
type ProviderStatus int

const (
	// providerStatusEnabled providers are enabled and have no limitations.
	providerStatusEnabled ProviderStatus = iota
	// providerStatusWarning providers are enabled, but show a warning message on manipulation.
	providerStatusWarning
	// providerStatusDisabled providers are disabled and cannot start or be manipulated.
	providerStatusDisabled
)

var errInvalidProviderStatus = errors.DefineInvalidArgument("invalid_provider_status", "invalid provider status `{status}`")

func providerStatusFromString(s string) (ProviderStatus, error) {
	switch s {
	case "enabled":
		return providerStatusEnabled, nil
	case "warning":
		return providerStatusWarning, nil
	case "disabled":
		return providerStatusDisabled, nil
	default:
		return ProviderStatus(0), errInvalidProviderStatus.WithAttributes("status", s)
	}
}

func providerTypeFromString(ctx context.Context, s string) (reflect.Type, error) {
	switch s {
	case "mqtt":
		return reflect.TypeOf(&ttnpb.ApplicationPubSub_Mqtt{}), nil
	case "nats":
		return reflect.TypeOf(&ttnpb.ApplicationPubSub_Nats{}), nil
	default:
		log.FromContext(ctx).WithField("provider", s).Warn("Unknown PubSub provider specified")
		return nil, nil
	}
}

// ProviderStatuses maps a provider type to a provider status.
type ProviderStatuses map[reflect.Type]ProviderStatus

// ProviderStatusesFromMap constructs the provider statuses from the provided map.
func ProviderStatusesFromMap(ctx context.Context, m map[string]string) (ProviderStatuses, error) {
	providers := make(ProviderStatuses)
	for k, v := range m {
		tp, err := providerTypeFromString(ctx, k)
		if err != nil {
			return nil, err
		}
		if tp == nil {
			continue
		}
		status, err := providerStatusFromString(v)
		if err != nil {
			return nil, err
		}
		providers[tp] = status
	}
	return providers, nil
}

var (
	errUnknownProvider  = errors.DefineInvalidArgument("unknown_provider", "provider `{provider}` is unknown")
	errProviderDisabled = errors.DefineFailedPrecondition("provider_disabled", "provider `{provider}` is disabled")
)

// Enabled checks if the provided provider is enabled.
// Providers which are not specified in the map are considered to be enabled by default.
func (ps ProviderStatuses) Enabled(ctx context.Context, provider ttnpb.ApplicationPubSub_Provider) error {
	tp := reflect.TypeOf(provider)
	name := strings.TrimPrefix(tp.String(), "*ttnpb.ApplicationPubSub_")
	switch ps[tp] {
	case providerStatusEnabled:
		return nil
	case providerStatusWarning:
		warning.Add(ctx, fmt.Sprintf("The %v Pub/Sub provider will be disabled in a future version of the stack", name))
		return nil
	case providerStatusDisabled:
		return errProviderDisabled.WithAttributes("provider", name)
	default:
		panic("unreachable pubsub provider status")
	}
}
