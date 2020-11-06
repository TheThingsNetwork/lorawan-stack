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
	"testing"

	"go.thethings.network/lorawan-stack/v3/pkg/component"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
)

// NewComponent returns a new Component that can be used for testing.
func NewComponent(tb testing.TB, config *component.Config, opts ...component.Option) *component.Component {
	c, err := component.New(test.GetLogger(tb), config, opts...)
	if err != nil {
		tb.Fatalf("Failed to create component: %v", err)
	}
	return c
}

// StartComponent starts the component for testing.
func StartComponent(tb testing.TB, c *component.Component) {
	if err := c.Start(); err != nil {
		tb.Fatalf("Failed to start component: %v", err)
	}
}
