// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package errors

import (
	"path/filepath"
	"runtime"
	"strings"
)

// rootPath is the root path of the project (usually go.thethings.network/lorawan-stack).
var rootPath string

func setRootPath() {
	pc, _, _, ok := runtime.Caller(0)
	if !ok {
		panic("could not determine import path of errors package")
	}
	fun := runtime.FuncForPC(pc).Name()
	rootPath = filepath.Dir(filepath.Dir(filepath.Join(filepath.Dir(fun), strings.Split(filepath.Base(fun), ".")[0]))) + "/"
}

func pkg(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		panic("could not determine source of error")
	}
	fun := runtime.FuncForPC(pc).Name()
	pkg := filepath.Join(filepath.Dir(fun), strings.Split(filepath.Base(fun), ".")[0])
	if rootPath == "" {
		setRootPath()
	}
	if strings.Contains(pkg, rootPath) {
		split := strings.Split(pkg, rootPath)
		pkg = split[len(split)-1]
	}
	return pkg
}
