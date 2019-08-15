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
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"gocloud.dev/gcerrors"
	"gocloud.dev/pubsub"
	"gocloud.dev/pubsub/driver"
)

type topic struct {
	client  mqtt.Client
	topic   string
	timeout time.Duration
	qos     byte
}

var errNilClient = errors.DefineInvalidArgument("nil_client", "client is nil")

// OpenTopic returns a *pubsub.Topic that publishes to the given topic name with the given MQTT client.
func OpenTopic(client mqtt.Client, topicName string, timeout time.Duration, qos byte) (*pubsub.Topic, error) {
	if client == nil {
		return nil, errNilClient
	}
	dt := &topic{
		client:  client,
		topic:   topicName,
		timeout: timeout,
		qos:     qos,
	}
	return pubsub.NewTopic(dt, nil), nil
}

// SendBatch implements driver.Topic.
func (t *topic) SendBatch(ctx context.Context, msgs []*driver.Message) error {
	for _, msg := range msgs {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if msg.BeforeSend != nil {
			asFunc := func(i interface{}) bool { return false }
			if err := msg.BeforeSend(asFunc); err != nil {
				return err
			}
		}
		if token := t.client.Publish(t.topic, t.qos, false, msg.Body); !token.WaitTimeout(t.timeout) {
			return token.Error()
		}
	}
	return nil
}

// IsRetryable implements driver.Topic.
func (*topic) IsRetryable(error) bool { return false }

// As implements driver.Topic.
func (t *topic) As(i interface{}) bool {
	c, ok := i.(*mqtt.Client)
	if !ok {
		return false
	}
	*c = t.client
	return true
}

// ErrorAs implements driver.Topic.
func (*topic) ErrorAs(error, interface{}) bool { return false }

// ErrorCode implements driver.Topic.
func (*topic) ErrorCode(err error) gcerrors.ErrorCode {
	return toErrorCode(err)
}

// Close implements driver.Topic.
func (*topic) Close() error { return nil }

type subscription struct {
	client  mqtt.Client
	topic   string
	subCh   chan mqtt.Message
	timeout time.Duration
}

// subscriptionQueueSize is the size of the subscription channel buffer.
const subscriptionQueueSize = 16

// OpenSubscription returns a *pubsub.Subscription that subscribes to the given topic name with the given MQTT client.
func OpenSubscription(client mqtt.Client, topicName string, timeout time.Duration, qos byte) (*pubsub.Subscription, error) {
	if client == nil {
		return nil, errNilClient
	}
	subCh := make(chan mqtt.Message, subscriptionQueueSize)
	handler := func(_ mqtt.Client, msg mqtt.Message) {
		subCh <- msg
	}
	if token := client.Subscribe(topicName, qos, handler); !token.WaitTimeout(timeout) {
		return nil, token.Error()
	}
	ds := &subscription{
		client:  client,
		topic:   topicName,
		subCh:   subCh,
		timeout: timeout,
	}
	return pubsub.NewSubscription(ds, nil, nil), nil
}

// ReceiveBatch implements driver.Subscription.
func (s *subscription) ReceiveBatch(ctx context.Context, maxMessages int) ([]*driver.Message, error) {
	var messages []*driver.Message
	for i := 0; i < maxMessages; i++ {
		select {
		case <-time.After(s.timeout):
			break
		case msg, ok := <-s.subCh:
			if !ok {
				break
			}
			messages = append(messages, &driver.Message{
				Body:  msg.Payload(),
				AckID: -1,
				AsFunc: func(i interface{}) bool {
					p, ok := i.(*mqtt.Message)
					if !ok {
						return false
					}
					*p = msg
					return true
				},
			})
		}
	}
	return messages, nil
}

// SendAcks implements driver.Subscription.
func (*subscription) SendAcks(context.Context, []driver.AckID) error { return nil }

// CanNack implements driver.Subscription.
func (*subscription) CanNack() bool { return false }

// SendNacks implements driver.Subscription.
func (*subscription) SendNacks(context.Context, []driver.AckID) error { panic("unreachable") }

// IsRetryable implements driver.Subscription.
func (*subscription) IsRetryable(error) bool { return false }

// As implements driver.Subscription.
func (s *subscription) As(i interface{}) bool {
	c, ok := i.(*mqtt.Client)
	if !ok {
		return false
	}
	*c = s.client
	return true
}

// ErrorAs implements driver.Subscription.
func (*subscription) ErrorAs(error, interface{}) bool { return false }

// ErrorCode implements driver.Subscription.
func (*subscription) ErrorCode(err error) gcerrors.ErrorCode {
	return toErrorCode(err)
}

// Close implements driver.Subscription.
func (s *subscription) Close() error {
	if token := s.client.Unsubscribe(s.topic); !token.WaitTimeout(s.timeout) {
		return token.Error()
	}
	return nil
}

func toErrorCode(err error) gcerrors.ErrorCode {
	switch err {
	case nil:
		return gcerrors.OK
	case context.Canceled:
		return gcerrors.Canceled
	case mqtt.ErrInvalidQos, mqtt.ErrInvalidTopicEmptyString, mqtt.ErrInvalidTopicMultilevel:
		return gcerrors.InvalidArgument
	case mqtt.ErrNotConnected:
		return gcerrors.NotFound
	default:
		return gcerrors.Unknown
	}
}
