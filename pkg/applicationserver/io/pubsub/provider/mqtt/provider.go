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

// Package mqtt implements the MQTT provider using the mqtt driver.
package mqtt

import (
	"context"
	"net/url"
	"time"

	mqtt_topic "github.com/TheThingsIndustries/mystique/pkg/topic"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/pubsub/provider"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"gocloud.dev/pubsub"
)

var timeout = (1 << 3) * time.Second

type impl struct {
}

type connection struct {
	mqtt.Client
}

// Shutdown implements provider.Shutdowner.
func (c *connection) Shutdown(_ context.Context) error {
	c.Disconnect(uint(timeout / time.Millisecond))
	return nil
}

var errConnectFailed = errors.Define("connect_failed", "connection to MQTT server failed")

// OpenConnection implements provider.Provider using the MQTT driver.
func (impl) OpenConnection(ctx context.Context, target provider.Target) (pc *provider.Connection, err error) {
	settings, ok := target.GetProvider().(*ttnpb.ApplicationPubSub_MQTT)
	if !ok {
		panic("wrong provider type provided to OpenConnection")
	}
	return OpenConnection(ctx, settings, target)
}

// OpenConnection opens a MQTT connection using the given settings.
func OpenConnection(ctx context.Context, settings *ttnpb.ApplicationPubSub_MQTT, topics provider.Topics) (*provider.Connection, error) {
	serverURL, err := adaptURLScheme(settings.MQTT.ServerURL)
	if err != nil {
		return nil, err
	}
	clientOpts := mqtt.NewClientOptions()
	clientOpts.AddBroker(serverURL)
	clientOpts.SetClientID(settings.MQTT.ClientID)
	clientOpts.SetUsername(settings.MQTT.Username)
	clientOpts.SetPassword(settings.MQTT.Password)
	if settings.MQTT.UseTLS {
		config, err := createTLSConfig(settings.MQTT.TLSCA, settings.MQTT.TLSClientCert, settings.MQTT.TLSClientKey)
		if err != nil {
			return nil, err
		}
		clientOpts.SetTLSConfig(config)
	}
	client := mqtt.NewClient(clientOpts)
	token := client.Connect()
	if !token.WaitTimeout(timeout) {
		return nil, errConnectFailed.WithCause(context.DeadlineExceeded)
	} else if token.Error() != nil {
		return nil, errConnectFailed.WithCause(token.Error())
	}
	pc := &provider.Connection{
		ProviderConnection: &connection{
			Client: client,
		},
	}
	for _, t := range []struct {
		topic   **pubsub.Topic
		message *ttnpb.ApplicationPubSub_Message
	}{
		{
			topic:   &pc.Topics.UplinkMessage,
			message: topics.GetUplinkMessage(),
		},
		{
			topic:   &pc.Topics.JoinAccept,
			message: topics.GetJoinAccept(),
		},
		{
			topic:   &pc.Topics.DownlinkAck,
			message: topics.GetDownlinkAck(),
		},
		{
			topic:   &pc.Topics.DownlinkNack,
			message: topics.GetDownlinkNack(),
		},
		{
			topic:   &pc.Topics.DownlinkSent,
			message: topics.GetDownlinkSent(),
		},
		{
			topic:   &pc.Topics.DownlinkFailed,
			message: topics.GetDownlinkFailed(),
		},
		{
			topic:   &pc.Topics.DownlinkQueued,
			message: topics.GetDownlinkQueued(),
		},
		{
			topic:   &pc.Topics.LocationSolved,
			message: topics.GetLocationSolved(),
		},
		{
			topic:   &pc.Topics.ServiceData,
			message: topics.GetServiceData(),
		},
	} {
		if t.message == nil {
			continue
		}
		if *t.topic, err = OpenTopic(
			client,
			mqtt_topic.Join(append(mqtt_topic.Split(topics.GetBaseTopic()), mqtt_topic.Split(t.message.GetTopic())...)),
			timeout,
			byte(settings.MQTT.PublishQoS),
		); err != nil {
			client.Disconnect(uint(timeout / time.Millisecond))
			return nil, err
		}
	}
	for _, s := range []struct {
		subscription **pubsub.Subscription
		message      *ttnpb.ApplicationPubSub_Message
	}{
		{
			subscription: &pc.Subscriptions.Push,
			message:      topics.GetDownlinkPush(),
		},
		{
			subscription: &pc.Subscriptions.Replace,
			message:      topics.GetDownlinkReplace(),
		},
	} {
		if s.message == nil {
			continue
		}
		if *s.subscription, err = OpenSubscription(
			client,
			mqtt_topic.Join(append(mqtt_topic.Split(topics.GetBaseTopic()), mqtt_topic.Split(s.message.GetTopic())...)),
			timeout,
			byte(settings.MQTT.SubscribeQoS),
		); err != nil {
			client.Disconnect(uint(timeout / time.Millisecond))
			return nil, err
		}
	}
	return pc, nil
}

func adaptURLScheme(initial string) (string, error) {
	u, err := url.Parse(initial)
	if err != nil {
		return "", err
	}
	switch u.Scheme {
	case "mqtt":
		u.Scheme = "tcp"
	case "mqtts":
		u.Scheme = "ssl"
	}
	return u.String(), nil
}

func init() {
	provider.RegisterProvider(&ttnpb.ApplicationPubSub_MQTT{}, impl{})
}
