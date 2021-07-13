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
	"crypto/tls"
	"net/http"
	"net/url"
	"time"

	mqtt_topic "github.com/TheThingsIndustries/mystique/pkg/topic"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/pubsub/provider"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"gocloud.dev/pubsub"
)

var shutdownTimeout = (1 << 3) * time.Second

type impl struct{}

type connection struct {
	mqtt.Client
}

// Shutdown implements provider.Shutdowner.
func (c *connection) Shutdown(_ context.Context) error {
	c.Disconnect(uint(shutdownTimeout / time.Millisecond))
	return nil
}

var errConnectFailed = errors.Define("connect_failed", "connection to MQTT server failed")

// OpenConnection implements provider.Provider using the MQTT driver.
func (impl) OpenConnection(ctx context.Context, target provider.Target, enabler provider.Enabler) (pc *provider.Connection, err error) {
	provider, ok := target.GetProvider().(*ttnpb.ApplicationPubSub_Mqtt)
	if !ok {
		panic("wrong provider type provided to OpenConnection")
	}
	if err := enabler.Enabled(ctx, target.GetProvider()); err != nil {
		return nil, err
	}

	var tlsConfig *tls.Config
	if provider.Mqtt.UseTls {
		var err error
		tlsConfig, err = createTLSConfig(provider.Mqtt.TlsCa, provider.Mqtt.TlsClientCert, provider.Mqtt.TlsClientKey)
		if err != nil {
			return nil, err
		}
	}

	headers := make(http.Header, len(provider.Mqtt.Headers))
	for k, v := range provider.Mqtt.Headers {
		headers.Set(k, v)
	}
	settings := Settings{
		URL:      provider.Mqtt.ServerUrl,
		ClientID: provider.Mqtt.ClientId,
		Username: provider.Mqtt.Username,
		Password: provider.Mqtt.Password,
		TLS:      tlsConfig,
		HTTPHeadersProvider: func(ctx context.Context) (http.Header, error) {
			return headers, nil
		},
		PublishQoS:   byte(provider.Mqtt.PublishQos),
		SubscribeQoS: byte(provider.Mqtt.SubscribeQos),
	}
	return OpenConnection(ctx, settings, target)
}

// HTTPHeadersProvider provides HTTP headers as they are needed by the MQTT client.
type HTTPHeadersProvider func(context.Context) (headers http.Header, err error)

// Settings configure the MQTT client.
type Settings struct {
	URL,
	ClientID,
	Username,
	Password string
	TLS                 *tls.Config
	HTTPHeadersProvider HTTPHeadersProvider
	PublishQoS,
	SubscribeQoS byte
	WebSocketReadBufferSize,
	WebSocketWriteBufferSize int
}

var errConfigureHTTPHeaders = errors.Define("configure_http_headers", "configure HTTP headers")

// OpenConnection opens a MQTT connection using the given settings.
func OpenConnection(ctx context.Context, settings Settings, topics provider.Topics) (pc *provider.Connection, err error) {
	serverURL, err := adaptURLScheme(settings.URL)
	if err != nil {
		return nil, err
	}
	logger := log.FromContext(ctx).WithFields(log.Fields(
		"url", serverURL,
		"client_id", settings.ClientID,
		"username", settings.Username,
	))

	clientOpts := mqtt.NewClientOptions()
	clientOpts.AddBroker(serverURL)
	clientOpts.SetClientID(settings.ClientID)
	clientOpts.SetUsername(settings.Username)
	clientOpts.SetPassword(settings.Password)
	clientOpts.SetTLSConfig(settings.TLS)
	clientOpts.SetKeepAlive(time.Minute)
	clientOpts.SetWebsocketOptions(&mqtt.WebsocketOptions{
		ReadBufferSize:  settings.WebSocketReadBufferSize,
		WriteBufferSize: settings.WebSocketWriteBufferSize,
	})

	if settings.HTTPHeadersProvider != nil {
		headers, err := settings.HTTPHeadersProvider(ctx)
		if err != nil {
			return nil, errConfigureHTTPHeaders.WithCause(err)
		}
		clientOpts.SetHTTPHeaders(headers)
	}
	clientOpts.SetOnConnectHandler(func(_ mqtt.Client) {
		logger.Info("Connected to MQTT server")
	})
	clientOpts.SetConnectionLostHandler(func(_ mqtt.Client, err error) {
		logger.WithError(err).Info("Disconnected from MQTT server")
	})
	clientOpts.SetReconnectingHandler(func(_ mqtt.Client, clientOpts *mqtt.ClientOptions) {
		logger.Info("Reconnect to MQTT server")
		if settings.HTTPHeadersProvider != nil {
			headers, err := settings.HTTPHeadersProvider(ctx)
			if err != nil {
				logger.WithError(err).Warn("Failed to configure HTTP headers on MQTT reconnect")
				return
			}
			clientOpts.SetHTTPHeaders(headers)
		}
	})

	client := mqtt.NewClient(clientOpts)
	token := client.Connect()
	defer func() {
		if err != nil {
			client.Disconnect(uint(shutdownTimeout / time.Millisecond))
		}
	}()
	if err := waitToken(ctx, token); err != nil {
		return nil, errConnectFailed.WithCause(err)
	}
	pc = &provider.Connection{
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
			topic:   &pc.Topics.DownlinkQueueInvalidated,
			message: topics.GetDownlinkQueueInvalidated(),
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
			settings.PublishQoS,
		); err != nil {
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
			ctx,
			client,
			mqtt_topic.Join(append(mqtt_topic.Split(topics.GetBaseTopic()), mqtt_topic.Split(s.message.GetTopic())...)),
			settings.SubscribeQoS,
		); err != nil {
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
	provider.RegisterProvider(&ttnpb.ApplicationPubSub_Mqtt{}, impl{})
}
