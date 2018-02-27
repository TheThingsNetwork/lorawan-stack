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

package translator_test

import (
	"fmt"

	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/udp"
	"github.com/TheThingsNetwork/ttn/pkg/gatewayserver/udp-translator"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

var ids = ttnpb.GatewayIdentifiers{GatewayID: "test-gateway"}

func downlinks() <-chan *ttnpb.GatewayDown { return make(chan *ttnpb.GatewayDown) }

func uplinks() chan<- *ttnpb.GatewayUp { return make(chan *ttnpb.GatewayUp) }

func getIP() string { return "127.0.0.1" }

func sendDownlinkToPF(udp.Data)     {}
func receiveUplinkFromPF() udp.Data { return udp.Data{} }

func Example() {
	t := translator.New(log.Noop)

	metadata := translator.Metadata{IP: getIP(), ID: ttnpb.GatewayIdentifiers{GatewayID: "My-Gateway-ID"}}

	go func() {
		for {
			select {
			case down, ok := <-downlinks():
				if !ok {
					return
				}

				output, err := t.Downlink(down)
				if err != nil {
					fmt.Println("Could no convert downlink to Semtech UDP format:", err)
					continue
				}

				sendDownlinkToPF(output)
			}
		}
	}()

	for {
		uplinkData := receiveUplinkFromPF()

		uplink, err := t.Upstream(uplinkData, metadata)
		if err != nil {
			fmt.Println("Could not convert uplink from Semtech UDP format:", err)
			continue
		}

		uplinks() <- uplink
	}
}
