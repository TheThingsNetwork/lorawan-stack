// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package frequencyplans

import (
	"path/filepath"
)

// DefaultListFilename is the default name that the frequency plans list has.
const DefaultListFilename = "frequency-plans.yml"

type storeReadConfiguration struct {
	Root string
}

func (c storeReadConfiguration) AbsolutePath(filename string) string {
	if filepath.IsAbs(filename) {
		return filename
	}

	if c.Root != "" {
		return filepath.Join(c.Root, filename)
	}

	return filename
}

// ReadFileSystemStoreOption is an option applied when creating the store from the filesystem.
type ReadFileSystemStoreOption func(*storeReadConfiguration)

// FileSystemRootPathOption can be used to specify the path to the directory where frequency plans will be read. The default path used is the directory of execution.
func FileSystemRootPathOption(path string) ReadFileSystemStoreOption {
	return func(config *storeReadConfiguration) {
		config.Root = path
	}
}
