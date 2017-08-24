// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package rpclog implements a gRPC logging middleware.
package rpclog

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/pkg/log"
	"google.golang.org/grpc/grpclog"
)

// ReplaceGrpcLogger sets the given log.Interface as a gRPC-level logger.
// This should be called *before* any other initialization, preferably from init() functions.
func ReplaceGrpcLogger(logger log.Interface) {
	zgl := &ttnGrpcLogger{logger}
	grpclog.SetLogger(zgl)
}

type ttnGrpcLogger struct {
	logger log.Interface
}

func (l *ttnGrpcLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(fmt.Sprint(args...))
}

func (l *ttnGrpcLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatal(fmt.Sprintf(format, args...))
}

func (l *ttnGrpcLogger) Fatalln(args ...interface{}) {
	l.logger.Fatal(fmt.Sprint(args...))
}

func (l *ttnGrpcLogger) Print(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}

func (l *ttnGrpcLogger) Printf(format string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(format, args...))
}

func (l *ttnGrpcLogger) Println(args ...interface{}) {
	l.logger.Info(fmt.Sprint(args...))
}
