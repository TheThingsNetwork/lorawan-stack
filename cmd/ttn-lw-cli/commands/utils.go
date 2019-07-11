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
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
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

var (
	utilitiesCommand = &cobra.Command{
		Use:               "utilities",
		Aliases:           []string{"utility", "util"},
		Short:             "Utilities (EXPERIMENTAL)",
		Hidden:            true,
		PersistentPreRunE: preRun(),
	}
	utilitiesBase64ToHexCommand = &cobra.Command{
		Use:     "base64-to-hex",
		Aliases: []string{"base64tohex"},
		Short:   "Convert base64 to hex",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			for _, arg := range args {
				var b []byte
				if len(arg)%4 == 0 {
					b, err = base64.StdEncoding.DecodeString(arg)
				} else {
					b, err = base64.RawStdEncoding.DecodeString(arg)
				}
				if err != nil {
					return err
				}
				fmt.Println(hex.EncodeToString(b))
			}
			return nil
		},
	}
	utilitiesHexToBase64Command = &cobra.Command{
		Use:     "hex-to-base64",
		Aliases: []string{"hextobase64"},
		Short:   "Convert hex to base64",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			for _, arg := range args {
				b, err := hex.DecodeString(strings.Replace(strings.TrimPrefix(arg, "0x"), " ", "", -1))
				if err != nil {
					return err
				}
				fmt.Println(base64.StdEncoding.EncodeToString(b))
			}
			return nil
		},
	}
)

func init() {
	Root.AddCommand(utilitiesCommand)
	utilitiesCommand.AddCommand(utilitiesBase64ToHexCommand)
	utilitiesCommand.AddCommand(utilitiesHexToBase64Command)
}
