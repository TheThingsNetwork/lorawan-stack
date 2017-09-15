// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package frequencyplans

type storeReadConfiguration struct {
	ListPath string
	Root     string
}

// ReadFilesystemStoreOption is an option applied when creating the store from the filesystem.
type ReadFilesystemStoreOption func(*storeReadConfiguration)

// FilesystemListPathOption can be used to specify the path to the list of frequency plans. The default path used is `frequency-plans.yml`, which is the path within the repository.
//
// If the path is a relative path, the file is read from the current directory, or from the root directory if it is specified.
func FilesystemListPathOption(path string) ReadFilesystemStoreOption {
	return func(config *storeReadConfiguration) {
		config.ListPath = path
	}
}

// ReadFilesystemStoreOption is an option applied when creating the store from the filesystem.
type ReadFilesystemStoreOption func(*storeReadConfiguration)

// FilesystemRootPathOption can be used to specify the path to the directory where frequency plans will be read. The default path used is the directory of execution.
func FilesystemRootPathOption(path string) ReadFilesystemStoreOption {
	return func(config *storeReadConfiguration) {
		config.Root = path
	}
}
