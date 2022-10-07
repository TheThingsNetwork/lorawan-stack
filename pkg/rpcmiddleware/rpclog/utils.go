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

package rpclog

import (
	"context"
	"fmt"
	"path"
	"strings"

	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	"go.thethings.network/lorawan-stack/v3/pkg/auth"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func logFieldsForError(err error) *log.F {
	fields := log.Fields()
	if ttnErr, ok := errors.From(err); ok {
		fields = fields.WithField("grpc_code", codes.Code(ttnErr.Code()))
		if ns := ttnErr.Namespace(); ns != "" {
			fields = fields.WithField("error_namespace", ns)
		}
		if name := ttnErr.Name(); name != "" {
			fields = fields.WithField("error_name", name)
		}
		if cid := ttnErr.CorrelationID(); cid != "" {
			fields = fields.WithField("error_correlation_id", cid)
		}
	} else if status, ok := status.FromError(err); ok {
		fields = fields.WithField("grpc_code", status.Code())
	}
	return fields
}

func logFieldsForCall(ctx context.Context, fullMethodString string) (once, propagated *log.F) {
	service := path.Dir(fullMethodString)[1:]
	method := path.Base(fullMethodString)

	once = log.Fields()
	propagated = log.Fields(
		"grpc.service", service,
		"grpc.method", method,
	)

	if ctxFields := grpc_ctxtags.Extract(ctx).Values(); len(ctxFields) > 0 {
		once = once.With(ctxFields)
	}

	if md, ok := metadata.FromOutgoingContext(ctx); ok {
		if requestID := md["x-request-id"]; len(requestID) > 0 {
			propagated = propagated.WithField("request_id", requestID[0])
		}
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if requestID := md["x-request-id"]; len(requestID) > 0 {
			propagated = propagated.WithField("request_id", requestID[0])
		}
		if xRealIP := md["x-real-ip"]; len(xRealIP) > 0 {
			propagated = propagated.WithField("peer.real_ip", xRealIP[0])
		}
		if authorization, ok := md["authorization"]; ok && len(authorization) > 0 {
			parts := strings.SplitN(authorization[len(authorization)-1], " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				if tokenType, tokenID, _, err := auth.SplitToken(parts[1]); err == nil {
					once = once.WithFields(log.Fields(
						"auth.token_type", tokenType.String(),
						"auth.token_id", tokenID,
					))
				}
			}
		}
	}

	return once, propagated
}

func commit(i log.Interface, level log.Level, msg string) {
	switch level {
	case log.DebugLevel:
		i.Debug(msg)
	case log.InfoLevel:
		i.Info(msg)
	case log.WarnLevel:
		i.Warn(msg)
	case log.ErrorLevel:
		i.Error(msg)
	case log.FatalLevel:
		i.Fatal(msg)
	default:
		panic(fmt.Sprintf("rpclog: unknown log level %d", level))
	}
}

func parseMethodLogCfg(opt string) (string, methodLogConfig) {
	optParts := strings.SplitN(opt, ":", 2)
	methodName := optParts[0]
	if len(optParts) == 1 {
		return methodName, methodLogConfig{IgnoreSuccess: true}
	}
	ignoredErrors := strings.Split(optParts[1], ";")
	isSuccessIgnored := ignoredErrors[0] == ""
	if isSuccessIgnored {
		ignoredErrors = ignoredErrors[1:]
	}
	ignoredErrorSet := make(map[string]struct{}, len(ignoredErrors))
	for _, ignoredMethodError := range ignoredErrors {
		ignoredErrorSet[ignoredMethodError] = struct{}{}
	}
	return methodName, methodLogConfig{
		IgnoreSuccess: isSuccessIgnored,
		IgnoredErrors: ignoredErrorSet,
	}
}
