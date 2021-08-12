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

// Package sentry implements a pkg/log.Handler that sends errors to Sentry
package sentry

import (
	"fmt"
	"strings"

	"github.com/getsentry/sentry-go"
	sentryerrors "go.thethings.network/lorawan-stack/v3/pkg/errors/sentry"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
)

// Sentry is a log.Handler that sends errors to Sentry.
type Sentry struct{}

// New creates a new Sentry log middleware.
func New() log.Middleware {
	return &Sentry{}
}

// Wrap an existing log handler with Sentry.
func (s *Sentry) Wrap(next log.Handler) log.Handler {
	return log.HandlerFunc(func(entry log.Entry) (err error) {
		if entry.Level() == log.ErrorLevel {
			s.forward(entry)
		}
		err = next.HandleLog(entry)
		return
	})
}

func (s *Sentry) forward(e log.Entry) *sentry.EventID {
	fields := e.Fields().Fields()

	switch fields["namespace"] {
	case "grpc", "web": // gRPC and web have their own Sentry integration.
		return nil
	}

	var err error
	if errField, ok := fields["error"]; ok {
		if errField, ok := errField.(error); ok {
			err = errField
		}
	}
	evt := sentryerrors.NewEvent(err)
	evt.Message = e.Message()
	if err == nil {
		evt.Level = sentry.LevelError
	}

	evt.Contexts["log fields"] = fields

	for k, v := range fields {
		switch k {
		case "namespace":
			evt.Tags["log.namespace"] = fmt.Sprint(v)
		case "auth.user_id":
			evt.User.ID = fmt.Sprint(v)
		case "peer.real_ip":
			evt.User.IPAddress = fmt.Sprint(v)
		case "grpc.service", "grpc.method",
			"auth.token_type", "auth.token_id":
			evt.Tags[k] = fmt.Sprint(v)
		case "dev_addr", "mac_version", "phy_version":
			evt.Tags[k] = fmt.Sprint(v)
		default:
			if strings.HasSuffix(k, "_id") || strings.HasSuffix(k, "_uid") || strings.HasSuffix(k, "_eui") {
				if strings.HasPrefix(k, "grpc.request.") {
					k = strings.TrimPrefix(k, "grpc.request.")
				}
				evt.Tags[k] = fmt.Sprint(v)
			}
		}
	}

	return sentry.CaptureEvent(evt)
}
