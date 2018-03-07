// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package fetch

import "fmt"

func Example() {
	fetcher := FromGitHubRepository("TheThingsNetwork/info", "master", "", true)
	content, err := fetcher.File("README.md")
	if err != nil {
		panic(err)
	}

	fmt.Println("Content of the README.md in TheThingsNetwork/info:")
	fmt.Println(string(content))
}
