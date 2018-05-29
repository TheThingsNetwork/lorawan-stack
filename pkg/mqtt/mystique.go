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

package mqtt

import (
	mqttlog "github.com/TheThingsIndustries/mystique/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/log"
)

type logWrapper struct {
	log.Interface
}

func (m logWrapper) WithField(str string, v interface{}) mqttlog.Interface {
	return logWrapper{
		Interface: m.Interface.WithField(str, v),
	}
}

func (m logWrapper) WithFields(f mqttlog.Fielder) mqttlog.Interface {
	return logWrapper{
		Interface: m.Interface.WithFields(f),
	}
}

func (m logWrapper) WithError(err error) mqttlog.Interface {
	return logWrapper{
		Interface: m.Interface.WithError(err),
	}
}

// Logger wraps a log.Interface to the Mystique logger interface.
func Logger(i log.Interface) mqttlog.Interface {
	return logWrapper{Interface: i}
}
