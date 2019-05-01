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
	"os"
	"sort"
	"strconv"
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
	colorTerm, _ := strconv.ParseBool(os.Getenv("COLORTERM"))
	level := log.InfoLevel
	if testing.Verbose() {
		level = log.DebugLevel
	}
	logger, err := log.NewLogger(
		log.WithLevel(level),
		log.WithHandler(log.NewCLI(os.Stdout, log.UseColor(colorTerm))),
	)
	if err != nil {
		t.Fatalf("Could not get logger: %v", err)
	}
	return &testLogger{
		stack:     logger,
		Interface: logger.WithField("test_name", t.Name()),
	}
}

type testLogger struct {
	stack *log.Logger
	log.Interface
}

func (l *testLogger) Use(m log.Middleware) {
	l.stack.Use(m)
}
