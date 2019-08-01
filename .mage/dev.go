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

package ttnmage

import (
	"os"
	"path"
	"runtime"

	"github.com/magefile/mage/mg"
)

// Dev namespace.
type Dev mg.Namespace

// Certificates generates certificates for development.
func (Dev) Certificates() error {
	if _, err := os.Stat("key.pem"); err == nil {
		if _, err := os.Stat("cert.pem"); err == nil {
			return nil
		}
	}
	return execGo("run", path.Join(runtime.GOROOT(), "src", "crypto", "tls", "generate_cert.go"), "-ca", "-host", "localhost,*.localhost")
}

func init() {
	initDeps = append(initDeps, Dev.Certificates)
}
