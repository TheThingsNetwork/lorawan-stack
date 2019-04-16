// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

// AuthCache stores auth for the CLI.
type AuthCache struct {
	data struct {
		OAuthToken *oauth2.Token          `json:"oauth_token,omitempty"`
		APIKey     string                 `json:"api_key,omitempty"`
		Hosts      []string               `json:"hosts,omitempty"`
		Other      map[string]interface{} `json:"other,omitempty"`
	}
	changed bool
}

// Get gets a key from the auth cache.
func (c *AuthCache) Get(key string) interface{} {
	switch key {
	case "oauth_token":
		return c.data.OAuthToken
	case "api_key":
		return c.data.APIKey
	case "hosts":
		return c.data.Hosts
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

// Set sets keys in the auth cache.
func (c *AuthCache) Set(key string, value interface{}) {
	switch key {
	case "oauth_token":
		c.data.OAuthToken = value.(*oauth2.Token)
	case "api_key":
		c.data.APIKey = value.(string)
	case "hosts":
		c.data.Hosts = value.([]string)
	default:
		if c.data.Other == nil {
			c.data.Other = make(map[string]interface{})
		}
		setInMap(c.data.Other, strings.Split(key, "."), value)
	}
	c.changed = true
}

// Unset unsets keys in the auth cache.
func (c *AuthCache) Unset(keys ...string) {
	for _, key := range keys {
		c.unset(key)
	}
}

func (c *AuthCache) unset(key string) {
	switch key {
	case "oauth_token":
		c.data.OAuthToken = nil
	case "api_key":
		c.data.APIKey = ""
	case "hosts":
		c.data.Hosts = nil
	default:
		if c.data.Other == nil {
			return
		}
		setInMap(c.data.Other, strings.Split(key, "."), nil)
	}
	c.changed = true
}

func setInMap(m map[string]interface{}, path []string, value interface{}) {
	if len(path) == 1 {
		if value == nil {
			delete(m, path[0])
		} else {
			m[path[0]] = value
		}
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

// GetAuthCache gets the auth cache form the cache file.
func GetAuthCache() (cache AuthCache, err error) {
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
	if err = json.NewDecoder(f).Decode(&cache.data); err != nil {
		return cache, err
	}
	return cache, nil
}

// SaveAuthCache saves the auth cache to the cache file.
func SaveAuthCache(cache AuthCache) (err error) {
	if !cache.changed {
		return nil
	}
	cacheFile := cacheFile()
	if cacheFile == "" {
		return nil
	}
	_, err = os.Stat(filepath.Dir(cacheFile))
	if os.IsNotExist(err) {
		err = os.MkdirAll(filepath.Dir(cacheFile), 0700)
	}
	if err != nil {
		return err
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
	return json.NewEncoder(f).Encode(&cache.data)
}
