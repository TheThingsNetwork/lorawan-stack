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

	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func logFieldsForError(err error) (fieldsKV []interface{}) {
	if ttnErr, ok := errors.From(err); ok {
		fieldsKV = append(fieldsKV, "grpc_code", codes.Code(ttnErr.Code()))
		if ns := ttnErr.Namespace(); ns != "" {
			fieldsKV = append(fieldsKV, "error_namespace", ns)
		}
		if name := ttnErr.Name(); name != "" {
			fieldsKV = append(fieldsKV, "error_name", name)
		}
		if cid := ttnErr.CorrelationID(); cid != "" {
			fieldsKV = append(fieldsKV, "error_correlation_id", cid)
		}
	} else if status, ok := status.FromError(err); ok {
		fieldsKV = append(fieldsKV,
			"grpc_code", status.Code(),
		)
	}
	return
}

func newLoggerForCall(ctx context.Context, logger log.Interface, fullMethodString string) context.Context {
	service := path.Dir(fullMethodString)[1:]
	method := path.Base(fullMethodString)
	if md, ok := metadata.FromOutgoingContext(ctx); ok {
		if requestID := md["request-id"]; len(requestID) > 0 {
			logger = logger.WithField("request_id", requestID[0])
		}
	}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if requestID := md["request-id"]; len(requestID) > 0 {
			logger = logger.WithField("request_id", requestID[0])
		}
	}
	return log.NewContext(ctx, logger.WithFields(log.Fields(
		"grpc_service", service,
		"grpc_method", method,
	)))
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
