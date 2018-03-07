// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package fetch

import "github.com/TheThingsNetwork/ttn/pkg/errors"

var (
	// ErrFileFailedToOpen indicates the file could not be opened.
	ErrFileFailedToOpen = &errors.ErrDescriptor{
		MessageFormat:  "File `{filename}` failed to open",
		Code:           1,
		Type:           errors.Internal,
		SafeAttributes: []string{"filename"},
	}
	// ErrFileNotFound indicates the file could not be found.
	ErrFileNotFound = &errors.ErrDescriptor{
		MessageFormat:  "File `{filename}` not found",
		Code:           2,
		Type:           errors.NotFound,
		SafeAttributes: []string{"filename"},
	}
)

func init() {
	ErrFileFailedToOpen.Register()
	ErrFileNotFound.Register()
}
