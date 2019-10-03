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

package log

import (
	"errors"
	"testing"
)

func TestLogInterface(t *testing.T) {
	oldDefault := Default
	defer func() { Default = oldDefault }()
	Default, _ = NewLogger(WithHandler(NoopHandler))

	Debug("test debug msg")
	Info("test info msg")
	Warn("test warn msg")
	Error("test error msg")
	Debugf("test debugf msg %d", 42)
	Infof("test infof msg %d", 42)
	Warnf("test warnf msg %d", 42)
	Errorf("test errorf msg %d", 42)

	for _, li := range []Interface{
		&Logger{},
		&noop{},
		WithField("key", "value"),
		WithFields(Fields("key", "value", "number", 42)),
		WithError(errors.New("unknown error")),
	} {
		for _, dli := range []Interface{
			li,
			li.WithField("key", "value"),
			li.WithFields(Fields("key", "value", "number", 42)),
			li.WithError(errors.New("unknown error")),
		} {
			dli.Debug("test debug msg")
			dli.Info("test info msg")
			dli.Warn("test warn msg")
			dli.Error("test error msg")
			dli.Debugf("test debugf msg %d", 42)
			dli.Infof("test infof msg %d", 42)
			dli.Warnf("test warnf msg %d", 42)
			dli.Errorf("test errorf msg %d", 42)
		}
	}
}
