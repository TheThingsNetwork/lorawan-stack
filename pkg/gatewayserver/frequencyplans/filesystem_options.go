// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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

// ReadFilesystemStoreOption is an option applied when creating the store from the filesystem.
type ReadFilesystemStoreOption func(*storeReadConfiguration)

// FilesystemRootPathOption can be used to specify the path to the directory where frequency plans will be read. The default path used is the directory of execution.
func FilesystemRootPathOption(path string) ReadFilesystemStoreOption {
	return func(config *storeReadConfiguration) {
		config.Root = path
	}
}
