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

package cluster_test

import (
	"context"
	"errors"
	"fmt"

	"go.thethings.network/lorawan-stack/pkg/cluster"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/rpcmiddleware/hooks"
)

func ExampleCluster() {
	c, err := cluster.New(context.Background(), &config.ServiceBase{
		// ...
	})
	if err != nil {
		panic(err)
	}

	hooks.RegisterUnaryHook("/ttn.lorawan.v3.Gs", cluster.HookName, c.Hook())
}

func ExampleIdentified() {
	handlingFunction := func(ctx context.Context, payload []byte) error {
		if !cluster.Identified(ctx) {
			return errors.New("The message did not come from the cluster")
		}

		return nil
	}

	err := handlingFunction(context.Background(), []byte("Hello world"))
	fmt.Println(err)
	// Output: The message did not come from the cluster
}
