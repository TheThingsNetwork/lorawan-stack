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

package api

import (
	"bytes"
	"encoding/json"
	"net/http"

	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/packages/lora-cloud-device-management-v1/api/objects"
)

// Tokens is an API client for the Token Management API.
type Tokens struct {
	cl *Client
}

const (
	tokenEntity = "token"
	nameParam   = "name"
	renewParam  = "renew"
)

// List returns the tokens.
func (t *Tokens) List() ([]objects.TokenInfo, error) {
	req, err := t.cl.newRequest(http.MethodGet, tokenEntity, "", listOperation, nil)
	if err != nil {
		return nil, err
	}
	resp, err := t.cl.cl.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var response tokensListResponse
	err = parse(&response, resp.Body)
	if err != nil {
		return nil, err
	}
	return response.Tokens, nil
}

// Update changes the name of the given token and optionally regenerates it.
func (t *Tokens) Update(token, newName string, renew bool) (*objects.TokenInfo, error) {
	var params []queryParam
	if newName != "" {
		params = append(params, queryParam{nameParam, newName})
	}
	if renew {
		params = append(params, queryParam{renewParam, ""})
	}
	resp, err := t.cl.Do(http.MethodPut, tokenEntity, token, updateOperation, nil, params...)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var response objects.TokenInfo
	err = parse(&response, resp.Body)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Add creates a token with the given name and capabilities.
func (t *Tokens) Add(name string, capabilities ...string) (*objects.TokenInfo, error) {
	buffer := bytes.NewBuffer(nil)
	err := json.NewEncoder(buffer).Encode(&tokenAddRequest{
		Name:         name,
		Capabilities: capabilities,
	})
	if err != nil {
		return nil, err
	}
	resp, err := t.cl.Do(http.MethodPost, tokenEntity, "", addOperation, buffer)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var response objects.TokenInfo
	err = parse(&response, resp.Body)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// Remove removes the given token.
func (t *Tokens) Remove(token string) error {
	resp, err := t.cl.Do(http.MethodDelete, tokenEntity, token, "", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	err = parse(nil, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

// Get returns the contents of the given token.
func (t *Tokens) Get(token string) (*objects.TokenInfo, error) {
	resp, err := t.cl.Do(http.MethodGet, tokenEntity, token, "", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var response objects.TokenInfo
	err = parse(&response, resp.Body)
	if err != nil {
		return nil, err
	}
	return &response, nil
}
