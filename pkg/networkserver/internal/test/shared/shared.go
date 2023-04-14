// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

// Package test contains testing utilities usable by all subpackages of networkserver excluding itself.
package test

import (
	"time"

	. "go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
)

var (
	CacheTTL           = (1 << 6) * test.Delay
	DefaultMACSettings = test.Must(DefaultConfig.DefaultMACSettings.Parse()).(*ttnpb.MACSettings)
)

type TaskPopFuncResponse struct {
	Time  time.Time
	Error error
}
