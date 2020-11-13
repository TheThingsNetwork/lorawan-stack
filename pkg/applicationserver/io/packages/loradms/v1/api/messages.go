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
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/golang/protobuf/proto"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
)

const sendOperation = "send"

type apiErrorDetail string

func (s *apiErrorDetail) Reset()        { *s = "" }
func (s apiErrorDetail) String() string { return string(s) }
func (apiErrorDetail) ProtoMessage()    {}

type baseResponse struct {
	Result interface{} `json:"result"`
	Errors []string    `json:"errors"`
}

var (
	errAPICallFailed = errors.Define("api_call_failed", "", "")
	errRequest       = errors.DefineUnavailable("request", "request failed with status `{code}`")
)

func parse(result interface{}, res *http.Response) error {
	defer res.Body.Close()
	defer io.Copy(ioutil.Discard, res.Body)
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return errRequest.WithAttributes("code", res.StatusCode)
	}
	r := &baseResponse{
		Result: result,
	}
	err := json.NewDecoder(res.Body).Decode(r)
	if err != nil {
		return err
	}
	if len(r.Errors) == 0 {
		return nil
	}
	var details []proto.Message
	for _, message := range r.Errors {
		ed := apiErrorDetail(message)
		details = append(details, &ed)
	}
	return errAPICallFailed.WithDetails(details...)
}
