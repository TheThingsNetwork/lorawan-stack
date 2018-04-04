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

package fetch

import (
	"fmt"
)

type gitHubBaseURLBuilder struct {
	branch     string
	repository string
	basePath   string
}

func (b gitHubBaseURLBuilder) render() string {
	url := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s", b.repository, b.branch)
	if b.basePath != "" {
		url = fmt.Sprintf("%s/%s", url, b.basePath)
	}
	return url
}

// FromGitHubRepository creates an Interface that fetches data from a GitHub repository.
//
// - repository: GitHub repository name in the TheThingsNetwork/ttn format
// - branch: Repository branch to use
// - basePath: Base path to look for the files
func FromGitHubRepository(repository, branch, basePath string, cache bool) Interface {
	builder := gitHubBaseURLBuilder{
		branch:     branch,
		basePath:   basePath,
		repository: repository,
	}

	return FromHTTP(builder.render(), cache)
}
