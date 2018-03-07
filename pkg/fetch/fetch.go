// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package fetch offers abstractions to fetch a file from a location (filesystem, git, HTTP...)
package fetch

// Interface is an abstraction for file retrieval.
type Interface interface {
	File(path string) ([]byte, error)
}
