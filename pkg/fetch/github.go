// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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
