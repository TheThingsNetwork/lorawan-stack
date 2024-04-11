// Copyright © 2024 The Things Network Foundation, The Things Industries B.V.
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

package sink

import (
	"io"
	"net/http"
	"unicode/utf8"

	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web/internal"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/structpb"
)

const maxResponseSize = (1 << 10) // 1 KiB

// httpClientSink contains an HTTP client to make outgoing requests.
type httpClientSink struct {
	*http.Client
}

var errRequest = errors.DefineUnavailable("request", "request", "webhook_id", "url", "status_code")

func requestErrorDetails(req *http.Request, res *http.Response) ([]any, []proto.Message) {
	ctx := req.Context()
	attributes, details := []any{
		"webhook_id", internal.WebhookIDFromContext(ctx).WebhookId,
		"url", req.URL.String(),
	}, []proto.Message{}
	if res != nil {
		attributes = append(attributes, "status_code", res.StatusCode)
		if body, _ := io.ReadAll(io.LimitReader(res.Body, maxResponseSize)); utf8.Valid(body) {
			details = append(details, &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"body": structpb.NewStringValue(string(body)),
				},
			})
		}
	}
	return attributes, details
}

func (s *httpClientSink) process(req *http.Request) error {
	res, err := s.Do(req)
	if err != nil {
		attributes, details := requestErrorDetails(req, res)
		return errRequest.WithAttributes(attributes...).WithDetails(details...).WithCause(err)
	}
	defer res.Body.Close()
	defer io.Copy(io.Discard, res.Body) //nolint:errcheck
	if res.StatusCode >= 200 && res.StatusCode <= 299 {
		return nil
	}
	attributes, details := requestErrorDetails(req, res)
	return errRequest.WithAttributes(attributes...).WithDetails(details...)
}

// Process uses the HTTP client to perform the requests.
func (s *httpClientSink) Process(req *http.Request) error {
	ctx := req.Context()
	if err := s.process(req); err != nil {
		registerWebhookFailed(ctx, err, true)
		return err
	}
	registerWebhookSent(ctx)
	return nil
}

// NewHTTPClientSink returns a new HTTP client sink.
func NewHTTPClientSink(client *http.Client) Sink {
	return &httpClientSink{client}
}
