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

package commands

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

var (
	errStorageIntegrationNotAvailable = errors.DefineUnimplemented("storage_integration_not_available", "Storage Integration not available")

	storageDBCommand = &cobra.Command{
		Use:   "storage-db",
		Short: "Manage the Storage Integration database",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("The storage integration is not available in the open-source version of The Things Stack.")
			fmt.Println("For more information, see https://www.thethingsindustries.com/docs/integrations/storage/")
			return errStorageIntegrationNotAvailable.New()
		},
	}
)

func init() {
	Root.AddCommand(storageDBCommand)
}
