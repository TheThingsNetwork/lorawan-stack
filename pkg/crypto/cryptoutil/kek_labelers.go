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

package cryptoutil

import (
	"context"
	"net"
	"net/url"
	"strings"

	"go.thethings.network/lorawan-stack/pkg/types"
)

// ComponentPrefixKEKLabeler is a ComponentKEKLabeler that joins the component prefix, separator and host.
type ComponentPrefixKEKLabeler struct {
	Separator string
}

func hostFromAddr(addr string) string {
	host := addr
	if url, err := url.Parse(addr); err == nil && url.Host != "" {
		host = url.Host
	}
	if h, _, err := net.SplitHostPort(host); err == nil {
		return h
	}
	return host
}

func (c ComponentPrefixKEKLabeler) join(parts ...string) string {
	sep := c.Separator
	if sep == "" {
		sep = ":"
	}
	return strings.Join(parts, sep)
}

// NsKEKLabel returns a KEK label in the form `ns:netID:host` from the given NetID and address, where `:` is the default separator. Empty parts are omitted.
func (c ComponentPrefixKEKLabeler) NsKEKLabel(ctx context.Context, netID *types.NetID, addr string) string {
	parts := make([]string, 0, 3)
	parts = append(parts, "ns")
	if netID != nil {
		parts = append(parts, netID.String())
	}
	if addr != "" {
		parts = append(parts, hostFromAddr(addr))
	}
	return c.join(parts...)
}

// AsKEKLabel returns a KEK label in the form `as:host` from the given address, where `:` is the default separator. Empty parts are omitted.
func (c ComponentPrefixKEKLabeler) AsKEKLabel(ctx context.Context, addr string) string {
	parts := make([]string, 0, 2)
	parts = append(parts, "as")
	if addr != "" {
		parts = append(parts, hostFromAddr(addr))
	}
	return c.join(parts...)
}
