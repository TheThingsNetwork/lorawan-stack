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
	"fmt"
	"os"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// Styl namespace.
type Styl mg.Namespace

func (styl Styl) stylint() (func(args ...string) (string, error), error) {
	if _, err := os.Stat(nodeBin("stylint")); os.IsNotExist(err) {
		if err = devDeps(); err != nil {
			return nil, err
		}
	}
	return func(args ...string) (string, error) {
		return sh.Output(nodeBin("stylint"), args...)
	}, nil
}

// Lint runs eslint over frontend js files.
func (styl Styl) Lint() error {
	if mg.Verbose() {
		fmt.Println("Running stylint")
	}
	stylint, err := styl.stylint()
	if err != nil {
		return err
	}
	res, err := stylint("./pkg/webui", "--config", "config/stylintrc.json")

	if res != "" {
		fmt.Println(res)
	}

	return err
}
