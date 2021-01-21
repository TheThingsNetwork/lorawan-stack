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

package devicerepository

import (
	"go.thethings.network/lorawan-stack/v3/pkg/devicerepository"
	"go.thethings.network/lorawan-stack/v3/pkg/devicerepository/store/bleve"
)

// DefaultDeviceRepositoryConfig is the default configuration for the Device Repository.
var DefaultDeviceRepositoryConfig = devicerepository.Config{
	Source:    "directory",
	Directory: "data/lorawan-devices",

	Store: devicerepository.StoreConfig{
		Bleve: bleve.Config{
			SearchPaths: []string{"lorawan-devices-index", "/srv/ttn-lorawan/lorawan-devices-index"},
		},
	},

	AssetsBaseURL: "https://raw.githubusercontent.com/TheThingsNetwork/lorawan-devices/master",
}
