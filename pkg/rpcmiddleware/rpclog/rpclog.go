// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

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
	grpclog.SetLoggerV2(&ttnGrpcLogger{logger})
}

type ttnGrpcLogger struct {
	logger log.Interface
}

func (l *ttnGrpcLogger) Info(args ...interface{})                    { l.logger.Debug(fmt.Sprint(args...)) }
func (l *ttnGrpcLogger) Infoln(args ...interface{})                  { l.logger.Debug(fmt.Sprint(args...)) }
func (l *ttnGrpcLogger) Infof(format string, args ...interface{})    { l.logger.Debugf(format, args...) }
func (l *ttnGrpcLogger) Warning(args ...interface{})                 { l.logger.Warn(fmt.Sprint(args...)) }
func (l *ttnGrpcLogger) Warningln(args ...interface{})               { l.logger.Warn(fmt.Sprint(args...)) }
func (l *ttnGrpcLogger) Warningf(format string, args ...interface{}) { l.logger.Warnf(format, args...) }
func (l *ttnGrpcLogger) Error(args ...interface{})                   { l.logger.Error(fmt.Sprint(args...)) }
func (l *ttnGrpcLogger) Errorln(args ...interface{})                 { l.logger.Error(fmt.Sprint(args...)) }
func (l *ttnGrpcLogger) Errorf(format string, args ...interface{})   { l.logger.Errorf(format, args...) }
func (l *ttnGrpcLogger) Fatal(args ...interface{})                   { l.logger.Fatal(fmt.Sprint(args...)) }
func (l *ttnGrpcLogger) Fatalln(args ...interface{})                 { l.logger.Fatal(fmt.Sprint(args...)) }
func (l *ttnGrpcLogger) Fatalf(format string, args ...interface{})   { l.logger.Fatalf(format, args...) }
func (l *ttnGrpcLogger) V(int) bool                                  { return true } // TODO: Use when log.Interface supports this
