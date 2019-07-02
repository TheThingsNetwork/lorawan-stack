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

package test_test

import (
	"errors"
	"testing"

	"go.thethings.network/lorawan-stack/pkg/log"
	. "go.thethings.network/lorawan-stack/pkg/util/test"
)

func TestGetLogger(t *testing.T) {
	var logger log.Interface = GetLogger(t)

	logger = logger.WithField("foo", "bar")

	logger = logger.WithError(errors.New("example error"))

	logger = logger.WithFields(log.Fields("k1", "v1", "k2", "v2"))

	logger.Debug("This is a debug log")
	logger.Info("This is a info log")
	logger.Warn("This is a warn log")
	logger.Error("This is an error log")

	logger.Debugf("This is a %s log", "debug")
	logger.Infof("This is a %s log", "info")
	logger.Warnf("This is a %s log", "warn")
	logger.Errorf("This is an %s log", "error")
}
