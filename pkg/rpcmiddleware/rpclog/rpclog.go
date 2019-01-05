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

// Package rpclog implements a gRPC logging middleware.
package rpclog

import (
	"fmt"
	"strings"

	"go.thethings.network/lorawan-stack/pkg/log"
	"google.golang.org/grpc/grpclog"
)

// ReplaceGrpcLogger sets the given log.Interface as a gRPC-level logger.
// This should be called *before* any other initialization, preferably from init() functions.
func ReplaceGrpcLogger(logger log.Interface) {
	grpclog.SetLoggerV2(&ttnGrpcLogger{logger})
}

var infoFilteredPrefixes = []string{
	`parsed scheme: ""`,
	`scheme "" not registered`,
	`ccResolverWrapper: sending new addresses`,
	`ClientConn switching balancer`,
}

type ttnGrpcLogger struct {
	logger log.Interface
}

func (l *ttnGrpcLogger) info(msg string) {
	for _, prefix := range infoFilteredPrefixes {
		if strings.HasPrefix(msg, prefix) {
			return
		}
	}
	l.logger.Debug(msg)
}
func (l *ttnGrpcLogger) Info(args ...interface{}) {
	l.info(fmt.Sprint(args...))
}
func (l *ttnGrpcLogger) Infoln(args ...interface{}) {
	l.info(fmt.Sprint(args...))
}
func (l *ttnGrpcLogger) Infof(format string, args ...interface{}) {
	l.info(fmt.Sprintf(format, args...))
}
func (l *ttnGrpcLogger) Warning(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}
func (l *ttnGrpcLogger) Warningln(args ...interface{}) {
	l.logger.Warn(fmt.Sprint(args...))
}
func (l *ttnGrpcLogger) Warningf(format string, args ...interface{}) {
	l.logger.Warnf(format, args...)
}
func (l *ttnGrpcLogger) Error(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}
func (l *ttnGrpcLogger) Errorln(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...))
}
func (l *ttnGrpcLogger) Errorf(format string, args ...interface{}) {
	l.logger.Errorf(format, args...)
}
func (l *ttnGrpcLogger) Fatal(args ...interface{}) {
	l.logger.Fatal(fmt.Sprint(args...))
}
func (l *ttnGrpcLogger) Fatalln(args ...interface{}) {
	l.logger.Fatal(fmt.Sprint(args...))
}
func (l *ttnGrpcLogger) Fatalf(format string, args ...interface{}) {
	l.logger.Fatalf(format, args...)
}
func (l *ttnGrpcLogger) V(int) bool {
	return true
} // TODO: Use when log.Interface supports this
