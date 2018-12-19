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

package util

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"
)

type Cache struct {
	data struct {
		OAuthToken *oauth2.Token          `json:"oauth_token,omitempty"`
		Other      map[string]interface{} `json:"other,omitempty"`
	}
	changed bool
}

func (c *Cache) Get(key string) interface{} {
	switch key {
	case "oauth_token":
		return c.data.OAuthToken
	default:
		return getFromMap(c.data.Other, strings.Split(key, "."))
	}
}

func getFromMap(m map[string]interface{}, path []string) interface{} {
	item := m[path[0]]
	if len(path) == 1 {
		return item
	}
	if m, ok := item.(map[string]interface{}); ok {
		return getFromMap(m, path[1:])
	}
	return nil
}

func (c *Cache) Set(key string, value interface{}) {
	switch key {
	case "oauth_token":
		c.data.OAuthToken = value.(*oauth2.Token)
	default:
		if c.data.Other == nil {
			c.data.Other = make(map[string]interface{})
		}
		setInMap(c.data.Other, strings.Split(key, "."), value)
	}
	c.changed = true
}

func setInMap(m map[string]interface{}, path []string, value interface{}) {
	if len(path) == 1 {
		m[path[0]] = value
	}
	if m[path[0]] == nil {
		m[path[0]] = make(map[string]interface{})
	}
	if m, ok := m[path[0]].(map[string]interface{}); ok {
		setInMap(m, path[1:], value)
	}
}

func cacheFile() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(cacheDir, "ttn-lw-cli", "cache")
}

func GetCache() (cache Cache, err error) {
	cacheFile := cacheFile()
	if cacheFile == "" {
		return cache, nil
	}
	f, err := os.OpenFile(cacheFile, os.O_RDONLY, 0)
	if err != nil {
		if os.IsNotExist(err) {
			return cache, nil
		}
		return cache, err
	}
	defer f.Close() // ignore errors
	err = json.NewDecoder(f).Decode(&cache.data)
	if err != nil {
		return cache, err
	}
	return cache, nil
}

func SaveCache(cache Cache) (err error) {
	if !cache.changed {
		return nil
	}
	cacheFile := cacheFile()
	if cacheFile == "" {
		return nil
	}
	f, err := os.OpenFile(cacheFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			err = f.Close()
		}
		if err != nil {
			os.Remove(cacheFile)
		}
	}()
	err = json.NewEncoder(f).Encode(&cache.data)
	if err != nil {
		return err
	}
	return nil
}
