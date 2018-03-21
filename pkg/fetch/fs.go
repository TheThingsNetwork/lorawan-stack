// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package fetch

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
)

type fsFetcher struct {
	basePath string
}

// FromFilesystem returns an interface that fetches files from the local filesystem
func FromFilesystem(basePath string) Interface {
	return fsFetcher{basePath: basePath}
}

func (f fsFetcher) File(pathElements ...string) ([]byte, error) {
	allElements := append([]string{f.basePath}, pathElements...)
	content, err := ioutil.ReadFile(filepath.Join(allElements...))
	if err == nil {
		return content, nil
	}

	attributes := errors.Attributes{
		"filename": filepath.Join(pathElements...),
	}
	switch err := err.(type) {
	case *os.PathError:
		if errno, ok := err.Err.(syscall.Errno); ok && errno == syscall.ENOENT {
			return nil, ErrFileNotFound.New(attributes)
		}
		return nil, ErrFileFailedToOpen.NewWithCause(attributes, err)
	default:
		return nil, err
	}
}
