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

package mqtt_test

import (
	"errors"
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/mqtt"
)

// recorder is log.Handler that records the entries
type recorder struct {
	entries []log.Entry
}

// HandleLog implements Handler
func (r *recorder) HandleLog(e log.Entry) error {
	r.entries = append(r.entries, e)
	return nil
}

func TestLogger(t *testing.T) {
	a := assertions.New(t)

	rec := &recorder{entries: []log.Entry{}}
	logger := &log.Logger{
		Level:   log.InfoLevel,
		Handler: rec,
	}
	mystiqueLogger := mqtt.Logger(logger)

	mystiqueLogger.Info("Yo!")
	a.So(rec.entries, should.HaveLength, 1)

	mystiqueLogger.WithField("test", "hi").Warn("Oops")
	a.So(rec.entries, should.HaveLength, 2)
	a.So(rec.entries[1].Fields().Fields()["test"], should.Equal, "hi")

	mystiqueLogger.WithError(errors.New("fatal")).Warn("Failure")
	a.So(rec.entries, should.HaveLength, 3)
	err, isError := rec.entries[2].Fields().Fields()["error"].(error)
	a.So(isError, should.BeTrue)
	a.So(err.Error(), should.Equal, "fatal")
}
