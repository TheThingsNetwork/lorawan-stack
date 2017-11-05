// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package rpclog

import (
	"fmt"
	"path"

	"context"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/grpc-ecosystem/go-grpc-middleware/tags"
)

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
	return log.WithLogger(ctx, logger.WithFields(log.Fields(
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
