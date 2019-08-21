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
	"time"

	mqtt_topic "github.com/TheThingsIndustries/mystique/pkg/topic"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/pubsub/provider"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"gocloud.dev/pubsub"
)

const defaultTimeout = (1 << 8) * time.Millisecond

type impl struct {
}

type connection struct {
	mqtt.Client
}

// Shutdown implements provider.Shutdowner.
func (c *connection) Shutdown(_ context.Context) error {
	c.Disconnect(uint(defaultTimeout / time.Millisecond))
	return nil
}

// OpenConnection implements provider.Provider using the mqtt driver.
func (impl) OpenConnection(ctx context.Context, pb *ttnpb.ApplicationPubSub) (pc *provider.Connection, err error) {
	if _, ok := pb.Provider.(*ttnpb.ApplicationPubSub_MQTT); !ok {
		panic("wrong provider type provided to OpenConnection")
	}
	settings := pb.GetMQTT()
	clientOpts := mqtt.NewClientOptions()
	clientOpts.AddBroker(settings.GetServerURL())
	clientOpts.SetClientID(settings.GetClientID())
	clientOpts.SetUsername(settings.GetUsername())
	clientOpts.SetPassword(settings.GetPassword())
	if settings.GetUseTLS() {
		config, err := createTLSConfig(settings.GetTLSCA(), settings.GetTLSClientCert(), settings.GetTLSClientKey())
		if err != nil {
			return nil, err
		}
		clientOpts.SetTLSConfig(config)
	}
	client := mqtt.NewClient(clientOpts)
	if token := client.Connect(); !token.WaitTimeout(defaultTimeout) {
		return nil, convertToCancelled(token.Error())
	}
	pc = &provider.Connection{
		ProviderConnection: &connection{
			Client: client,
		},
	}
	for _, t := range []struct {
		topic     **pubsub.Topic
		topicName string
	}{
		{
			topic:     &pc.Topics.UplinkMessage,
			topicName: pb.GetUplinkMessage().GetTopic(),
		},
		{
			topic:     &pc.Topics.JoinAccept,
			topicName: pb.GetJoinAccept().GetTopic(),
		},
		{
			topic:     &pc.Topics.DownlinkAck,
			topicName: pb.GetDownlinkAck().GetTopic(),
		},
		{
			topic:     &pc.Topics.DownlinkNack,
			topicName: pb.GetDownlinkNack().GetTopic(),
		},
		{
			topic:     &pc.Topics.DownlinkSent,
			topicName: pb.GetDownlinkSent().GetTopic(),
		},
		{
			topic:     &pc.Topics.DownlinkFailed,
			topicName: pb.GetDownlinkFailed().GetTopic(),
		},
		{
			topic:     &pc.Topics.DownlinkQueued,
			topicName: pb.GetDownlinkQueued().GetTopic(),
		},
		{
			topic:     &pc.Topics.LocationSolved,
			topicName: pb.GetLocationSolved().GetTopic(),
		},
	} {
		if *t.topic, err = OpenTopic(
			client,
			mqtt_topic.Join(append(mqtt_topic.Split(pb.BaseTopic), mqtt_topic.Split(t.topicName)...)),
			defaultTimeout,
			byte(settings.GetPublishQoS()),
		); err != nil {
			client.Disconnect(uint(defaultTimeout / time.Millisecond))
			return nil, err
		}
	}
	for _, s := range []struct {
		subscription **pubsub.Subscription
		topicName    string
	}{
		{
			subscription: &pc.Subscriptions.Push,
			topicName:    pb.GetDownlinkPush().GetTopic(),
		},
		{
			subscription: &pc.Subscriptions.Replace,
			topicName:    pb.GetDownlinkReplace().GetTopic(),
		},
	} {
		if *s.subscription, err = OpenSubscription(
			client,
			mqtt_topic.Join(append(mqtt_topic.Split(pb.BaseTopic), mqtt_topic.Split(s.topicName)...)),
			defaultTimeout,
			byte(settings.GetSubscribeQoS()),
		); err != nil {
			client.Disconnect(uint(defaultTimeout / time.Millisecond))
			return nil, err
		}
	}
	return pc, nil
}

func init() {
	provider.RegisterProvider(&ttnpb.ApplicationPubSub_MQTT{}, impl{})
}
