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

package mqtt

import (
	"context"
	"fmt"
	"testing"
	"time"

	paho_mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"gocloud.dev/pubsub/driver"
	"gocloud.dev/pubsub/drivertest"
)

type harness struct {
	client paho_mqtt.Client
}

func (h *harness) CreateTopic(ctx context.Context, testName string) (dt driver.Topic, cleanup func(), err error) {
	dt, err = openDriverTopic(h.client, fmt.Sprintf("test/%s", testName), timeout, 1)
	if err != nil {
		return nil, func() {}, err
	}
	return dt, func() {}, err
}

func (h *harness) MakeNonexistentTopic(ctx context.Context) (driver.Topic, error) {
	return (*topic)(nil), nil
}

func (h *harness) CreateSubscription(ctx context.Context, t driver.Topic, testName string) (ds driver.Subscription, cleanup func(), err error) {
	dt, err := openDriverSubscription(h.client, fmt.Sprintf("test/%s", testName), timeout, 1)
	if err != nil {
		return nil, func() {}, err
	}
	return dt, func() {}, nil
}

func (h *harness) MakeNonexistentSubscription(ctx context.Context) (driver.Subscription, error) {
	return (*subscription)(nil), nil
}

func (h *harness) Close() {
	h.client.Disconnect(uint(timeout / time.Millisecond))
}

func (h *harness) MaxBatchSizes() (int, int) { return 1, 1 }

func (h *harness) SupportsMultipleSubscriptions() bool { return false }

func createHarnessMaker(broker string) func(context.Context, *testing.T) (drivertest.Harness, error) {
	return func(ctx context.Context, t *testing.T) (drivertest.Harness, error) {
		clientOpts := paho_mqtt.NewClientOptions()
		clientOpts.AddBroker(broker)
		client := paho_mqtt.NewClient(clientOpts)
		token := client.Connect()
		if !token.WaitTimeout(timeout) {
			t.Fatal("Connection timeout")
		}
		if err := token.Error(); err != nil {
			t.Fatal(err)
		}
		return &harness{client: client}, nil
	}
}

func TestConformance(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()
	a := assertions.New(t)

	lis, _, err := startMQTTServer(ctx, nil)
	a.So(err, should.BeNil)
	a.So(lis, should.NotBeNil)
	defer lis.Close()

	drivertest.RunConformanceTests(t, createHarnessMaker(fmt.Sprintf("tcp://%v", lis.Addr())), nil)
}

type mockMessage struct {
	paho_mqtt.Message
	payload []byte
}

func (m *mockMessage) Payload() []byte {
	return m.payload
}

func TestEncodeDecodeMessage(t *testing.T) {
	for _, tc := range []struct {
		name string
		dm   *driver.Message
		body []byte
	}{
		{
			name: "OnlyBody",
			dm: &driver.Message{
				Body:     []byte{0x01, 0x02, 0x03},
				Metadata: nil,
			},
		},
		{
			name: "OnlyMetadata",
			dm: &driver.Message{
				Body: nil,
				Metadata: map[string]string{
					"foo": "bar",
				},
			},
		},
		{
			name: "BodyAndMetadata",
			dm: &driver.Message{
				Body: []byte{0x01, 0x02, 0x03},
				Metadata: map[string]string{
					"foo": "bar",
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := assertions.New(t)

			body, err := encodeMessage(tc.dm)
			a.So(err, should.BeNil)

			dm, err := decodeMessage(&mockMessage{
				payload: body,
			})
			a.So(err, should.BeNil)
			a.So(dm, should.NotBeNil)
			a.So(dm.Body, should.Resemble, tc.dm.Body)
			a.So(dm.Metadata, should.Resemble, tc.dm.Metadata)
		})
	}
}
