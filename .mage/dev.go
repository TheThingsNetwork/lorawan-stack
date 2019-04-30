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
	"path"
	"runtime"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/target"
	errfmt "golang.org/x/exp/errors/fmt"
)

// Dev namespace.
type Dev mg.Namespace

// Certificates generates certificates for development.
func (Dev) Certificates() error {
	changed, err := target.Glob("{key,cert}.pem")
	if err != nil {
		return errfmt.Errorf("failed checking modtime: %w", err)
	}
	if !changed {
		return nil
	}
	return execGo("run", path.Join(runtime.GOROOT(), "src", "crypto", "tls", "generate_cert.go"), "-ca", "-host", "localhost")
}

func init() {
	initDeps = append(initDeps, Dev.Certificates)
}
