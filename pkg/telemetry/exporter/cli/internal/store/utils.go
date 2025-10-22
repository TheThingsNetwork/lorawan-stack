// Copyright © 2025 The Things Network Foundation, The Things Industries B.V.
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

package store

import (
	"os"
	"path"
)

// ttnPath returns the path to the folder in which the configuration of telemetry is stored.
func ttnPath() (string, error) {
	configPath, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return path.Join(configPath, "ttn-lw-cli"), nil
}

// dbPath returns the path to the file in which the CLI telemetry is stored.
func dbPath() (string, error) {
	p, err := ttnPath()
	if err != nil {
		return "", err
	}
	return path.Join(p, "telemetry.json"), nil
}
