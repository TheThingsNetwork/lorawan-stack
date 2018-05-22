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

package fs

import (
	"net/http"
	"os"
	"path"
)

// CombinedFileSystem is a http.FileSystem that tries two filesystems in a row for reading files.
type CombinedFileSystem struct {
	primary   http.FileSystem
	secondary http.FileSystem
}

// Open implements http.FileSystem.
func (c *CombinedFileSystem) Open(name string) (http.File, error) {
	if c.primary != nil {
		file, err := c.primary.Open(name)
		if err == nil {
			return file, nil
		}
	}

	return c.secondary.Open(name)
}

// Combine returns a new http.FileSystem that will first try the primary filesystem
// to read a file, otherwise then try the secondary filesystem if the file did not
// exist in the primary FileSystem.
func Combine(primary, secondary http.FileSystem) *CombinedFileSystem {
	return &CombinedFileSystem{
		primary:   primary,
		secondary: secondary,
	}
}

// SubDir is a filesystem that sits at a specific subdirectory of another filesystem.
type SubDir struct {
	fs     http.FileSystem
	prefix string
}

// Subdirectory returns a new filesystem rooted at the specified subdirectory of the filesystem.
func Subdirectory(fs http.FileSystem, dir string) *SubDir {
	return &SubDir{
		fs:     fs,
		prefix: dir,
	}
}

// Open implements http.FileSystem.
func (fs *SubDir) Open(name string) (http.File, error) {
	return fs.fs.Open(path.Join(fs.prefix, name))
}

// HiddenFilesystem is a filesystem that hides some files.
type HiddenFilesystem struct {
	fs     http.FileSystem
	hidden map[string]bool
}

// Hide creates a new filesystem that hides some files.
func Hide(fs http.FileSystem, files ...string) *HiddenFilesystem {
	hfs := &HiddenFilesystem{
		fs:     fs,
		hidden: make(map[string]bool),
	}
	for _, file := range files {
		hfs.hidden[file] = true
	}
	return hfs
}

// Open implements http.FileSystem.
func (fs *HiddenFilesystem) Open(name string) (http.File, error) {
	if _, hidden := fs.hidden[name]; hidden {
		return nil, os.ErrNotExist
	}

	return fs.fs.Open(name)
}
