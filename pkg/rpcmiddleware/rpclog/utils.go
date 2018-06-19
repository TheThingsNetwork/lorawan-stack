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

package rpclog

import (
	"context"
	"fmt"
	"path"

	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func logFieldsForError(err error) (fieldsKV []interface{}) {
	if err, ok := errors.From(err); ok {
		fieldsKV = append(fieldsKV, "grpc_code", codes.Code(err.Code()))
		if ns := err.Namespace(); ns != "" {
			fieldsKV = append(fieldsKV, "error_namespace", ns)
		}
		if name := err.Name(); name != "" {
			fieldsKV = append(fieldsKV, "error_name", name)
		}
		if cid := err.CorrelationID(); cid != "" {
			fieldsKV = append(fieldsKV, "error_correlation_id", cid)
		}
	} else if err, ok := status.FromError(err); ok {
		fieldsKV = append(fieldsKV,
			"grpc_code", err.Code(),
		)
	}
	return
}

type fielder struct {
	values map[string]interface{}
}

func (f fielder) Fields() map[string]interface{} {
	return f.values
}

func newLoggerForCall(ctx context.Context, logger log.Interface, fullMethodString string) context.Context {
	service := path.Dir(fullMethodString)[1:]
	method := path.Base(fullMethodString)
	if tags := grpc_ctxtags.Extract(ctx).Values(); len(tags) > 0 {
		logger = logger.WithFields(&fielder{values: tags})
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
