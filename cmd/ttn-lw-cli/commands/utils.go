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
	"io/ioutil"
	"net"
	"net/url"
	"strings"

	"github.com/spf13/pflag"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/io"
	"go.thethings.network/lorawan-stack/v3/cmd/ttn-lw-cli/internal/util"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
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

func payloadFormatterParameterFlags(prefix string) *pflag.FlagSet {
	flagSet := &pflag.FlagSet{}
	flagSet.AddFlagSet(dataFlags(prefix+".down-formatter-parameter", ""))
	flagSet.AddFlagSet(dataFlags(prefix+".up-formatter-parameter", ""))
	return flagSet
}

// parsePayloadFormatterParameterFlags parses formatter-parameter-local-file arguments,
// updates formatters with the file contents and returns the extra field mask paths.
func parsePayloadFormatterParameterFlags(prefix string, formatters *ttnpb.MessagePayloadFormatters, flags *pflag.FlagSet) ([]string, error) {
	if formatters == nil {
		return nil, nil
	}
	paths := []string{}
	r, err := getDataReader(prefix+".up-formatter-parameter", flags)
	switch err {
	case nil:
		b, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		formatters.UpFormatterParameter = string(b)
		paths = append(paths, prefix+".up-formatter-parameter")
	default:
		if !errors.IsInvalidArgument(err) {
			return nil, err
		}
	}

	r, err = getDataReader(prefix+".down-formatter-parameter", flags)
	switch err {
	case nil:
		b, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		formatters.DownFormatterParameter = string(b)
		paths = append(paths, prefix+".down-formatter-parameter")
	default:
		if !errors.IsInvalidArgument(err) {
			return nil, err
		}
	}
	return util.NormalizePaths(paths), nil
}
