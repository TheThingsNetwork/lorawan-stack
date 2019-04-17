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

// AuthData is the stored auth data.
type AuthData struct {
	OAuthToken *oauth2.Token          `json:"oauth_token,omitempty"`
	APIKey     string                 `json:"api_key,omitempty"`
	Hosts      []string               `json:"hosts,omitempty"`
	Other      map[string]interface{} `json:"other,omitempty"`
}

// AuthCache stores auth for the CLI.
type AuthCache struct {
	data struct {
		AuthData
		ByID map[string]*AuthData `json:"by_id"`
	}
	id      string
	changed bool
}

// ForID returns the auth cache for the given ID.
func (c AuthCache) ForID(id string) AuthCache {
	clone := c
	clone.id = id
	return clone
}

func (c *AuthCache) getData() *AuthData {
	data := &c.data.AuthData
	if c.id != "" {
		if c.data.ByID == nil {
			c.data.ByID = make(map[string]*AuthData)
		}
		var ok bool
		data, ok = c.data.ByID[c.id]
		if !ok {
			data = &AuthData{}
			c.data.ByID[c.id] = data
		}
	}
	return data
}

// Get gets a key from the auth cache.
func (c *AuthCache) Get(key string) interface{} {
	authData := c.getData()
	switch key {
	case "oauth_token":
		return authData.OAuthToken
	case "api_key":
		return authData.APIKey
	case "hosts":
		return authData.Hosts
	default:
		return getFromMap(authData.Other, strings.Split(key, "."))
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
	authData := c.getData()
	switch key {
	case "oauth_token":
		authData.OAuthToken = value.(*oauth2.Token)
	case "api_key":
		authData.APIKey = value.(string)
	case "hosts":
		authData.Hosts = value.([]string)
	default:
		if authData.Other == nil {
			authData.Other = make(map[string]interface{})
		}
		setInMap(authData.Other, strings.Split(key, "."), value)
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
	authData := c.getData()
	switch key {
	case "oauth_token":
		authData.OAuthToken = nil
	case "api_key":
		authData.APIKey = ""
	case "hosts":
		authData.Hosts = nil
	default:
		if authData.Other == nil {
			return
		}
		setInMap(authData.Other, strings.Split(key, "."), nil)
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
