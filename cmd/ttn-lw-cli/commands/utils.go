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

package commands

import (
	"fmt"
	stdio "io"
	"net"
	"net/url"
	"strings"

	"go.thethings.network/lorawan-stack/cmd/ttn-lw-cli/internal/io"
)

func getHost(address string) string {
	if strings.Contains(address, "://") {
		url, err := url.Parse(address)
		if err == nil {
			address = url.Host
		}
	}
	if strings.Contains(address, ":") {
		host, _, err := net.SplitHostPort(address)
		if err == nil {
			return host
		}
	}
	return address
}

func getHosts(addresses ...string) []string {
	hostmap := make(map[string]struct{})
	for _, address := range addresses {
		hostmap[getHost(address)] = struct{}{}
	}
	hosts := make([]string, 0, len(hostmap))
	for host := range hostmap {
		hosts = append(hosts, host)
	}
	return hosts
}

func getInputDecoder(reader stdio.Reader) (io.Decoder, error) {
	switch config.InputFormat {
	case "json":
		return io.NewJSONDecoder(reader), nil
	default:
		return nil, fmt.Errorf("unknown input format: %s", config.InputFormat)
	}
}
