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

package mac

import (
	"fmt"
	"unicode"

	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

func macEventOptions(extraOpts ...events.Option) []events.Option {
	return append([]events.Option{events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ)}, extraOpts...)
}

func defineReceiveMACAcceptEvent(name, desc string, opts ...events.Option) func() events.Builder {
	return events.DefineFunc(
		fmt.Sprintf("ns.mac.%s.answer.accept", name), fmt.Sprintf("%s accept received", desc),
		macEventOptions(opts...)...,
	)
}

func defineReceiveMACAnswerEvent(name, desc string, opts ...events.Option) func() events.Builder {
	return events.DefineFunc(
		fmt.Sprintf("ns.mac.%s.answer", name), fmt.Sprintf("%s answer received", desc),
		macEventOptions(opts...)...,
	)
}

func defineReceiveMACIndicationEvent(name, desc string, opts ...events.Option) func() events.Builder {
	return events.DefineFunc(
		fmt.Sprintf("ns.mac.%s.indication", name), fmt.Sprintf("%s indication received", desc),
		macEventOptions(opts...)...,
	)
}

func defineReceiveMACRejectEvent(name, desc string, opts ...events.Option) func() events.Builder {
	return events.DefineFunc(
		fmt.Sprintf("ns.mac.%s.answer.reject", name), fmt.Sprintf("%s rejection received", desc),
		macEventOptions(opts...)...,
	)
}

func defineReceiveMACRequestEvent(name, desc string, opts ...events.Option) func() events.Builder {
	return events.DefineFunc(
		fmt.Sprintf("ns.mac.%s.request", name), fmt.Sprintf("%s request received", desc),
		macEventOptions(opts...)...,
	)
}

func defineEnqueueMACAnswerEvent(name, desc string, opts ...events.Option) func() events.Builder {
	return events.DefineFunc(
		fmt.Sprintf("ns.mac.%s.answer", name), fmt.Sprintf("%s answer enqueued", desc),
		macEventOptions(opts...)...,
	)
}

func defineEnqueueMACConfirmationEvent(name, desc string, opts ...events.Option) func() events.Builder {
	return events.DefineFunc(
		fmt.Sprintf("ns.mac.%s.confirmation", name), fmt.Sprintf("%s confirmation enqueued", desc),
		macEventOptions(opts...)...,
	)
}

func defineEnqueueMACRequestEvent(name, desc string, opts ...events.Option) func() events.Builder {
	return events.DefineFunc(
		fmt.Sprintf("ns.mac.%s.request", name), fmt.Sprintf("%s request enqueued", desc),
		macEventOptions(opts...)...,
	)
}

func defineClassSwitchEvent(class rune) func() events.Builder {
	return events.DefineFunc(
		fmt.Sprintf("ns.class.switch.%c", class), fmt.Sprintf("switched to class %c", unicode.ToUpper(class)),
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
	)
}

var (
	EvtEnqueueProprietaryMACAnswer  = defineEnqueueMACAnswerEvent("proprietary", "proprietary MAC command")
	EvtEnqueueProprietaryMACRequest = defineEnqueueMACRequestEvent("proprietary", "proprietary MAC command")
	EvtReceiveProprietaryMAC        = events.Define(
		"ns.mac.proprietary.receive", "receive proprietary MAC command",
		events.WithVisibility(ttnpb.Right_RIGHT_APPLICATION_TRAFFIC_READ),
	)

	EvtClassASwitch = defineClassSwitchEvent('a')()
	EvtClassBSwitch = defineClassSwitchEvent('b')()
	EvtClassCSwitch = defineClassSwitchEvent('c')()
)
