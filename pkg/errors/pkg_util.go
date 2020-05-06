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

package errors

import (
	"runtime"
	"strings"
)

var pkgPrefix = func() string {
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		panic("could not determine package of pkg_util.go")
	}
	p := strings.TrimSuffix(runtime.FuncForPC(pc).Name(), "pkg/errors.init")
	return p
}()

// namespace is called when errors are defined.
// It returns the package path of the caller (skipping the first frames of the call stack)
// and makes it relative (so for example: pkg/errors).
func namespace(skip int) string {
	pc, _, _, ok := runtime.Caller(skip)
	if !ok {
		panic("could not determine source of error")
	}
	fun := runtime.FuncForPC(pc).Name()
	slashIdx := strings.LastIndexByte(fun, '/')
	dotIdx := strings.IndexByte(fun[slashIdx:], '.')
	pkg := fun[:slashIdx+dotIdx]
	return strings.TrimPrefix(pkg, pkgPrefix)
}
