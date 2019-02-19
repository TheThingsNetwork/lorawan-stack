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

package commands

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

var cancelSignals = []os.Signal{syscall.SIGHUP, os.Interrupt, syscall.SIGTERM}

func newContext(parent context.Context) context.Context {
	ctx, cancel := context.WithCancel(parent)
	sig := make(chan os.Signal)
	signal.Notify(sig, cancelSignals...)
	go func() {
		select {
		case <-ctx.Done():
		case sig := <-sig:
			logger.WithField("signal", sig).Debug("Command interrupted")
			cancel()
		}
		signal.Stop(sig)
	}()
	return ctx
}
