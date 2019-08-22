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

package ttnpb

import (
	"strconv"
)

// ApplicationPubSub_Provider is an alias to the interface identifying the PubSub provider types.
// This enables provider.RegisterProvider and provider.GetProvider to offer type safety guarantees.
// The underscore is maintained for consistency with the generated code.
type ApplicationPubSub_Provider = isApplicationPubSub_Provider

// MarshalText implements encoding.TextMarshaler interface.
func (q ApplicationPubSub_MQTTProvider_QoS) MarshalText() ([]byte, error) {
	return []byte(q.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler interface.
func (q *ApplicationPubSub_MQTTProvider_QoS) UnmarshalText(b []byte) error {
	s := string(b)
	if i, ok := ApplicationPubSub_MQTTProvider_QoS_value[s]; ok {
		*q = ApplicationPubSub_MQTTProvider_QoS(i)
		return nil
	}
	if i, err := strconv.Atoi(s); err == nil {
		if _, ok := ApplicationPubSub_MQTTProvider_QoS_name[int32(i)]; ok {
			*q = ApplicationPubSub_MQTTProvider_QoS(int32(i))
			return nil
		}
	}
	return errCouldNotParse("ApplicationPubSub_MQTTProvider_QoS")(string(b))
}
