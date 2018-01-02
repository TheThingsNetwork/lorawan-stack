// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package test

const success = ""

func all(results ...string) string {
	for _, res := range results {
		if res != success {
			return res
		}
	}

	return success
}
