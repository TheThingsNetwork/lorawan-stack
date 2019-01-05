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

package test

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"go.thethings.network/lorawan-stack/pkg/log"
)

var colorTerm = os.Getenv("COLORTERM") != "0"

type kv struct {
	k string
	v interface{}
}

type fields []kv

func (f fields) Len() int           { return len(f) }
func (f fields) Swap(i, j int)      { f[i], f[j] = f[j], f[i] }
func (f fields) Less(i, j int) bool { return f[i].k < f[j].k }

func (f fields) unique() fields {
	uniqueMap := make(map[string]interface{}, len(f))
	for _, kv := range f {
		uniqueMap[kv.k] = kv.v
	}
	unique := make(fields, 0, len(uniqueMap))
	for k, v := range uniqueMap {
		unique = append(unique, kv{k, v})
	}
	return unique
}

func (f fields) sorted() fields {
	clone := make(fields, len(f))
	copy(clone, f)
	sort.Sort(clone)
	return clone
}

// GetLogger returns a logger for tests.
func GetLogger(t testing.TB) log.Stack {
	return &testLogger{t: t}
}

type testLogger struct {
	t      testing.TB
	fields fields
}

func (l *testLogger) Use(m log.Middleware) {}

func (l *testLogger) format(level log.Level, msg string, v ...interface{}) string {
	if len(v) > 0 {
		msg = fmt.Sprintf(msg, v...)
	}
	levelStr := strings.ToUpper(level.String())
	if colorTerm {
		levelStr = fmt.Sprintf("\033[%dm%6s\033[0m", log.Colors[level], level)
	}
	formatted := fmt.Sprintf("%s %-40s", levelStr, msg)
	if len(l.fields) > 0 {
		for _, kv := range l.fields.unique().sorted() {
			key := kv.k
			if colorTerm {
				key = fmt.Sprintf(" \033[%dm%s\033[0m", log.Colors[level], kv.k)
			}
			formatted += fmt.Sprintf(" %s=%v", key, kv.v)
		}
	}
	return formatted
}

func (l *testLogger) Debug(msg string) {
	l.t.Helper()
	l.t.Log(l.format(log.DebugLevel, msg))
}
func (l *testLogger) Info(msg string) {
	l.t.Helper()
	l.t.Log(l.format(log.InfoLevel, msg))
}
func (l *testLogger) Warn(msg string) {
	l.t.Helper()
	l.t.Log(l.format(log.WarnLevel, msg))
}
func (l *testLogger) Error(msg string) {
	l.t.Helper()
	l.t.Log(l.format(log.ErrorLevel, msg))
}
func (l *testLogger) Fatal(msg string) {
	l.t.Helper()
	l.t.Log(l.format(log.FatalLevel, msg))
}
func (l *testLogger) Debugf(msg string, v ...interface{}) {
	l.t.Helper()
	l.t.Log(l.format(log.DebugLevel, msg, v...))
}
func (l *testLogger) Infof(msg string, v ...interface{}) {
	l.t.Helper()
	l.t.Log(l.format(log.InfoLevel, msg, v...))
}
func (l *testLogger) Warnf(msg string, v ...interface{}) {
	l.t.Helper()
	l.t.Log(l.format(log.WarnLevel, msg, v...))
}
func (l *testLogger) Errorf(msg string, v ...interface{}) {
	l.t.Helper()
	l.t.Log(l.format(log.ErrorLevel, msg, v...))
}
func (l *testLogger) Fatalf(msg string, v ...interface{}) {
	l.t.Helper()
	l.t.Log(l.format(log.FatalLevel, msg, v...))
}
func (l *testLogger) WithField(k string, v interface{}) log.Interface {
	return &testLogger{
		t:      l.t,
		fields: append(l.fields, kv{k, v}),
	}
}
func (l *testLogger) WithFields(f log.Fielder) log.Interface {
	m := f.Fields()
	fields := make([]kv, 0, len(m))
	for k, v := range m {
		fields = append(fields, kv{k, v})
	}
	return &testLogger{
		t:      l.t,
		fields: append(l.fields, fields...),
	}
}
func (l *testLogger) WithError(err error) log.Interface {
	return l.WithField("error", err)
}
