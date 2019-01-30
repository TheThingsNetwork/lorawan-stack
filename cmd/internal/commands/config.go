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
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/pkg/config"
)

// Config returns a command that prints the current configuration in the config manager.
func Config(mgr *config.Manager) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "View the current configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			space := 0
			for _, key := range mgr.AllKeys() {
				if len(key)+8 > space {
					space = len(key) + 8
				}
			}
			useEnv, _ := cmd.Flags().GetBool("env")
			joinSlice := func(s []string) string {
				if useEnv {
					return strings.Join(s, " ")
				}
				return strings.Join(s, ",")
			}
			for _, key := range mgr.AllKeys() {
				flagOrEnv, val := key, mgr.Get(key)
				if useEnv {
					flagOrEnv = mgr.EnvironmentForKey(flagOrEnv)
				} else {
					flagOrEnv = "--" + flagOrEnv
				}
				var empty bool
				switch v := val.(type) {
				case []string:
					if len(v) == 0 {
						empty = true
					} else {
						val = joinSlice(v)
					}
				case map[string]string:
					if len(v) == 0 {
						empty = true
					} else {
						var pairs []string
						for k, v := range v {
							pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
						}
						val = joinSlice(pairs)
					}
				case map[string][]uint8:
					if len(v) == 0 {
						empty = true
					} else {
						var pairs []string
						for k, v := range v {
							pairs = append(pairs, fmt.Sprintf("%s=%x", k, v))
						}
						val = joinSlice(pairs)
					}
				}
				if empty {
					continue
				}
				if useEnv {
					fmt.Fprintf(cmd.OutOrStdout(), "%s=\"%v\"\n", flagOrEnv, val)
				} else {
					fmt.Fprintf(cmd.OutOrStdout(), "%"+strconv.Itoa(space)+"s=\"%v\"\n", flagOrEnv, val)
				}
			}
			return nil
		},
	}
	cmd.Flags().Bool("env", false, "print as environment")
	return cmd
}
