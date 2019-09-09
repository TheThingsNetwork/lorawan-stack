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

package fetch

import "strings"

type basePathFetcher struct {
	Interface
	basePath []string
}

func (f basePathFetcher) File(pathElements ...string) ([]byte, error) {
	if len(pathElements) == 0 {
		return nil, errFilenameNotSpecified
	}
	if strings.HasPrefix(pathElements[0], "/") {
		return f.Interface.File(pathElements...)
	}
	return f.Interface.File(append(f.basePath, pathElements...)...)
}

func WithBasePath(f Interface, basePath ...string) Interface {
	return basePathFetcher{
		Interface: f,
		basePath:  append(basePath[:0:0], basePath...),
	}
}
