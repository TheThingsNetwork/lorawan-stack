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
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"google.golang.org/grpc/codes"
)

const (
	sendOperation = "send"

	maxResponseSize = (1 << 24)
)

type baseResponse struct {
	Result interface{} `json:"result"`
	Errors []string    `json:"errors"`
}

var errRequest = errors.Define("request", "LoRaCloud DMS request failed")

func parse(result interface{}, res *http.Response) error {
	defer res.Body.Close()
	defer io.Copy(ioutil.Discard, res.Body)
	reader := io.LimitReader(res.Body, maxResponseSize)
	if res.StatusCode < 200 || res.StatusCode > 299 {
		body, _ := ioutil.ReadAll(reader)
		return errRequest.WithDetails(&ttnpb.ErrorDetails{
			Code:          uint32(res.StatusCode),
			MessageFormat: string(body),
		})
	}
	r := &baseResponse{
		Result: result,
	}
	if err := json.NewDecoder(reader).Decode(r); err != nil {
		return err
	}
	if len(r.Errors) == 0 {
		return nil
	}
	var details []proto.Message
	for _, message := range r.Errors {
		details = append(details, &ttnpb.ErrorDetails{
			Code:          uint32(codes.Unknown),
			MessageFormat: message,
		})
	}
	return errRequest.WithDetails(details...)
}
