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
	"io"

	"github.com/spf13/cobra"
)

// asBulk enables some commands to do bulk operations.
// If there is a non-nil input decoder, asBulk keeps executing the same command
// until it returns an error (the input decoder returns io.EOF when it's done).
// If the input decoder is nil, the command is executed only once.
func asBulk(runE func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) (err error) {
		if inputDecoder == nil {
			return runE(cmd, args)
		}
		for {
			err = runE(cmd, args)
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
		}
	}
}
