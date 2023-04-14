// Code generated by protoc-gen-go-json. DO NOT EDIT.
// versions:
// - protoc-gen-go-json v1.5.1
// - protoc             v4.22.2
// source: lorawan-stack/api/applicationserver_pubsub.proto

package ttnpb

import (
	golang "github.com/TheThingsIndustries/protoc-gen-go-json/golang"
	jsonplugin "github.com/TheThingsIndustries/protoc-gen-go-json/jsonplugin"
)

// MarshalProtoJSON marshals the ApplicationPubSub_MQTTProvider_QoS to JSON.
func (x ApplicationPubSub_MQTTProvider_QoS) MarshalProtoJSON(s *jsonplugin.MarshalState) {
	s.WriteEnumString(int32(x), ApplicationPubSub_MQTTProvider_QoS_name)
}

// MarshalText marshals the ApplicationPubSub_MQTTProvider_QoS to text.
func (x ApplicationPubSub_MQTTProvider_QoS) MarshalText() ([]byte, error) {
	return []byte(jsonplugin.GetEnumString(int32(x), ApplicationPubSub_MQTTProvider_QoS_name)), nil
}

// MarshalJSON marshals the ApplicationPubSub_MQTTProvider_QoS to JSON.
func (x ApplicationPubSub_MQTTProvider_QoS) MarshalJSON() ([]byte, error) {
	return jsonplugin.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the ApplicationPubSub_MQTTProvider_QoS from JSON.
func (x *ApplicationPubSub_MQTTProvider_QoS) UnmarshalProtoJSON(s *jsonplugin.UnmarshalState) {
	v := s.ReadEnum(ApplicationPubSub_MQTTProvider_QoS_value)
	if err := s.Err(); err != nil {
		s.SetErrorf("could not read QoS enum: %v", err)
		return
	}
	*x = ApplicationPubSub_MQTTProvider_QoS(v)
}

// UnmarshalText unmarshals the ApplicationPubSub_MQTTProvider_QoS from text.
func (x *ApplicationPubSub_MQTTProvider_QoS) UnmarshalText(b []byte) error {
	i, err := jsonplugin.ParseEnumString(string(b), ApplicationPubSub_MQTTProvider_QoS_value)
	if err != nil {
		return err
	}
	*x = ApplicationPubSub_MQTTProvider_QoS(i)
	return nil
}

// UnmarshalJSON unmarshals the ApplicationPubSub_MQTTProvider_QoS from JSON.
func (x *ApplicationPubSub_MQTTProvider_QoS) UnmarshalJSON(b []byte) error {
	return jsonplugin.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

// MarshalProtoJSON marshals the ApplicationPubSub_MQTTProvider message to JSON.
func (x *ApplicationPubSub_MQTTProvider) MarshalProtoJSON(s *jsonplugin.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if x.ServerUrl != "" || s.HasField("server_url") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("server_url")
		s.WriteString(x.ServerUrl)
	}
	if x.ClientId != "" || s.HasField("client_id") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("client_id")
		s.WriteString(x.ClientId)
	}
	if x.Username != "" || s.HasField("username") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("username")
		s.WriteString(x.Username)
	}
	if x.Password != "" || s.HasField("password") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("password")
		s.WriteString(x.Password)
	}
	if x.SubscribeQos != 0 || s.HasField("subscribe_qos") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("subscribe_qos")
		x.SubscribeQos.MarshalProtoJSON(s)
	}
	if x.PublishQos != 0 || s.HasField("publish_qos") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("publish_qos")
		x.PublishQos.MarshalProtoJSON(s)
	}
	if x.UseTls || s.HasField("use_tls") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("use_tls")
		s.WriteBool(x.UseTls)
	}
	if len(x.TlsCa) > 0 || s.HasField("tls_ca") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("tls_ca")
		s.WriteBytes(x.TlsCa)
	}
	if len(x.TlsClientCert) > 0 || s.HasField("tls_client_cert") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("tls_client_cert")
		s.WriteBytes(x.TlsClientCert)
	}
	if len(x.TlsClientKey) > 0 || s.HasField("tls_client_key") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("tls_client_key")
		s.WriteBytes(x.TlsClientKey)
	}
	if x.Headers != nil || s.HasField("headers") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("headers")
		s.WriteObjectStart()
		var wroteElement bool
		for k, v := range x.Headers {
			s.WriteMoreIf(&wroteElement)
			s.WriteObjectStringField(k)
			s.WriteString(v)
		}
		s.WriteObjectEnd()
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the ApplicationPubSub_MQTTProvider to JSON.
func (x *ApplicationPubSub_MQTTProvider) MarshalJSON() ([]byte, error) {
	return jsonplugin.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the ApplicationPubSub_MQTTProvider message from JSON.
func (x *ApplicationPubSub_MQTTProvider) UnmarshalProtoJSON(s *jsonplugin.UnmarshalState) {
	if s.ReadNil() {
		return
	}
	s.ReadObject(func(key string) {
		switch key {
		default:
			s.ReadAny() // ignore unknown field
		case "server_url", "serverUrl":
			s.AddField("server_url")
			x.ServerUrl = s.ReadString()
		case "client_id", "clientId":
			s.AddField("client_id")
			x.ClientId = s.ReadString()
		case "username":
			s.AddField("username")
			x.Username = s.ReadString()
		case "password":
			s.AddField("password")
			x.Password = s.ReadString()
		case "subscribe_qos", "subscribeQos":
			s.AddField("subscribe_qos")
			x.SubscribeQos.UnmarshalProtoJSON(s)
		case "publish_qos", "publishQos":
			s.AddField("publish_qos")
			x.PublishQos.UnmarshalProtoJSON(s)
		case "use_tls", "useTls":
			s.AddField("use_tls")
			x.UseTls = s.ReadBool()
		case "tls_ca", "tlsCa":
			s.AddField("tls_ca")
			x.TlsCa = s.ReadBytes()
		case "tls_client_cert", "tlsClientCert":
			s.AddField("tls_client_cert")
			x.TlsClientCert = s.ReadBytes()
		case "tls_client_key", "tlsClientKey":
			s.AddField("tls_client_key")
			x.TlsClientKey = s.ReadBytes()
		case "headers":
			s.AddField("headers")
			if s.ReadNil() {
				x.Headers = nil
				return
			}
			x.Headers = make(map[string]string)
			s.ReadStringMap(func(key string) {
				x.Headers[key] = s.ReadString()
			})
		}
	})
}

// UnmarshalJSON unmarshals the ApplicationPubSub_MQTTProvider from JSON.
func (x *ApplicationPubSub_MQTTProvider) UnmarshalJSON(b []byte) error {
	return jsonplugin.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

// MarshalProtoJSON marshals the ApplicationPubSub message to JSON.
func (x *ApplicationPubSub) MarshalProtoJSON(s *jsonplugin.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if x.Ids != nil || s.HasField("ids") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("ids")
		// NOTE: ApplicationPubSubIdentifiers does not seem to implement MarshalProtoJSON.
		golang.MarshalMessage(s, x.Ids)
	}
	if x.CreatedAt != nil || s.HasField("created_at") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("created_at")
		if x.CreatedAt == nil {
			s.WriteNil()
		} else {
			golang.MarshalTimestamp(s, x.CreatedAt)
		}
	}
	if x.UpdatedAt != nil || s.HasField("updated_at") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("updated_at")
		if x.UpdatedAt == nil {
			s.WriteNil()
		} else {
			golang.MarshalTimestamp(s, x.UpdatedAt)
		}
	}
	if x.Format != "" || s.HasField("format") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("format")
		s.WriteString(x.Format)
	}
	if x.Provider != nil {
		switch ov := x.Provider.(type) {
		case *ApplicationPubSub_Nats:
			s.WriteMoreIf(&wroteField)
			s.WriteObjectField("nats")
			// NOTE: ApplicationPubSub_NATSProvider does not seem to implement MarshalProtoJSON.
			golang.MarshalMessage(s, ov.Nats)
		case *ApplicationPubSub_Mqtt:
			s.WriteMoreIf(&wroteField)
			s.WriteObjectField("mqtt")
			ov.Mqtt.MarshalProtoJSON(s.WithField("mqtt"))
		case *ApplicationPubSub_AwsIot:
			s.WriteMoreIf(&wroteField)
			s.WriteObjectField("aws_iot")
			// NOTE: ApplicationPubSub_AWSIoTProvider does not seem to implement MarshalProtoJSON.
			golang.MarshalMessage(s, ov.AwsIot)
		}
	}
	if x.BaseTopic != "" || s.HasField("base_topic") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("base_topic")
		s.WriteString(x.BaseTopic)
	}
	if x.DownlinkPush != nil || s.HasField("downlink_push") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("downlink_push")
		// NOTE: ApplicationPubSub_Message does not seem to implement MarshalProtoJSON.
		golang.MarshalMessage(s, x.DownlinkPush)
	}
	if x.DownlinkReplace != nil || s.HasField("downlink_replace") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("downlink_replace")
		// NOTE: ApplicationPubSub_Message does not seem to implement MarshalProtoJSON.
		golang.MarshalMessage(s, x.DownlinkReplace)
	}
	if x.UplinkMessage != nil || s.HasField("uplink_message") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("uplink_message")
		// NOTE: ApplicationPubSub_Message does not seem to implement MarshalProtoJSON.
		golang.MarshalMessage(s, x.UplinkMessage)
	}
	if x.UplinkNormalized != nil || s.HasField("uplink_normalized") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("uplink_normalized")
		// NOTE: ApplicationPubSub_Message does not seem to implement MarshalProtoJSON.
		golang.MarshalMessage(s, x.UplinkNormalized)
	}
	if x.JoinAccept != nil || s.HasField("join_accept") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("join_accept")
		// NOTE: ApplicationPubSub_Message does not seem to implement MarshalProtoJSON.
		golang.MarshalMessage(s, x.JoinAccept)
	}
	if x.DownlinkAck != nil || s.HasField("downlink_ack") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("downlink_ack")
		// NOTE: ApplicationPubSub_Message does not seem to implement MarshalProtoJSON.
		golang.MarshalMessage(s, x.DownlinkAck)
	}
	if x.DownlinkNack != nil || s.HasField("downlink_nack") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("downlink_nack")
		// NOTE: ApplicationPubSub_Message does not seem to implement MarshalProtoJSON.
		golang.MarshalMessage(s, x.DownlinkNack)
	}
	if x.DownlinkSent != nil || s.HasField("downlink_sent") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("downlink_sent")
		// NOTE: ApplicationPubSub_Message does not seem to implement MarshalProtoJSON.
		golang.MarshalMessage(s, x.DownlinkSent)
	}
	if x.DownlinkFailed != nil || s.HasField("downlink_failed") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("downlink_failed")
		// NOTE: ApplicationPubSub_Message does not seem to implement MarshalProtoJSON.
		golang.MarshalMessage(s, x.DownlinkFailed)
	}
	if x.DownlinkQueued != nil || s.HasField("downlink_queued") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("downlink_queued")
		// NOTE: ApplicationPubSub_Message does not seem to implement MarshalProtoJSON.
		golang.MarshalMessage(s, x.DownlinkQueued)
	}
	if x.DownlinkQueueInvalidated != nil || s.HasField("downlink_queue_invalidated") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("downlink_queue_invalidated")
		// NOTE: ApplicationPubSub_Message does not seem to implement MarshalProtoJSON.
		golang.MarshalMessage(s, x.DownlinkQueueInvalidated)
	}
	if x.LocationSolved != nil || s.HasField("location_solved") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("location_solved")
		// NOTE: ApplicationPubSub_Message does not seem to implement MarshalProtoJSON.
		golang.MarshalMessage(s, x.LocationSolved)
	}
	if x.ServiceData != nil || s.HasField("service_data") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("service_data")
		// NOTE: ApplicationPubSub_Message does not seem to implement MarshalProtoJSON.
		golang.MarshalMessage(s, x.ServiceData)
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the ApplicationPubSub to JSON.
func (x *ApplicationPubSub) MarshalJSON() ([]byte, error) {
	return jsonplugin.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the ApplicationPubSub message from JSON.
func (x *ApplicationPubSub) UnmarshalProtoJSON(s *jsonplugin.UnmarshalState) {
	if s.ReadNil() {
		return
	}
	s.ReadObject(func(key string) {
		switch key {
		default:
			s.ReadAny() // ignore unknown field
		case "ids":
			s.AddField("ids")
			if s.ReadNil() {
				x.Ids = nil
				return
			}
			// NOTE: ApplicationPubSubIdentifiers does not seem to implement UnmarshalProtoJSON.
			var v ApplicationPubSubIdentifiers
			golang.UnmarshalMessage(s, &v)
			x.Ids = &v
		case "created_at", "createdAt":
			s.AddField("created_at")
			if s.ReadNil() {
				x.CreatedAt = nil
				return
			}
			v := golang.UnmarshalTimestamp(s)
			if s.Err() != nil {
				return
			}
			x.CreatedAt = v
		case "updated_at", "updatedAt":
			s.AddField("updated_at")
			if s.ReadNil() {
				x.UpdatedAt = nil
				return
			}
			v := golang.UnmarshalTimestamp(s)
			if s.Err() != nil {
				return
			}
			x.UpdatedAt = v
		case "format":
			s.AddField("format")
			x.Format = s.ReadString()
		case "nats":
			s.AddField("nats")
			ov := &ApplicationPubSub_Nats{}
			x.Provider = ov
			if s.ReadNil() {
				ov.Nats = nil
				return
			}
			// NOTE: ApplicationPubSub_NATSProvider does not seem to implement UnmarshalProtoJSON.
			var v ApplicationPubSub_NATSProvider
			golang.UnmarshalMessage(s, &v)
			ov.Nats = &v
		case "mqtt":
			ov := &ApplicationPubSub_Mqtt{}
			x.Provider = ov
			if s.ReadNil() {
				ov.Mqtt = nil
				return
			}
			ov.Mqtt = &ApplicationPubSub_MQTTProvider{}
			ov.Mqtt.UnmarshalProtoJSON(s.WithField("mqtt", true))
		case "aws_iot", "awsIot":
			s.AddField("aws_iot")
			ov := &ApplicationPubSub_AwsIot{}
			x.Provider = ov
			if s.ReadNil() {
				ov.AwsIot = nil
				return
			}
			// NOTE: ApplicationPubSub_AWSIoTProvider does not seem to implement UnmarshalProtoJSON.
			var v ApplicationPubSub_AWSIoTProvider
			golang.UnmarshalMessage(s, &v)
			ov.AwsIot = &v
		case "base_topic", "baseTopic":
			s.AddField("base_topic")
			x.BaseTopic = s.ReadString()
		case "downlink_push", "downlinkPush":
			s.AddField("downlink_push")
			if s.ReadNil() {
				x.DownlinkPush = nil
				return
			}
			// NOTE: ApplicationPubSub_Message does not seem to implement UnmarshalProtoJSON.
			var v ApplicationPubSub_Message
			golang.UnmarshalMessage(s, &v)
			x.DownlinkPush = &v
		case "downlink_replace", "downlinkReplace":
			s.AddField("downlink_replace")
			if s.ReadNil() {
				x.DownlinkReplace = nil
				return
			}
			// NOTE: ApplicationPubSub_Message does not seem to implement UnmarshalProtoJSON.
			var v ApplicationPubSub_Message
			golang.UnmarshalMessage(s, &v)
			x.DownlinkReplace = &v
		case "uplink_message", "uplinkMessage":
			s.AddField("uplink_message")
			if s.ReadNil() {
				x.UplinkMessage = nil
				return
			}
			// NOTE: ApplicationPubSub_Message does not seem to implement UnmarshalProtoJSON.
			var v ApplicationPubSub_Message
			golang.UnmarshalMessage(s, &v)
			x.UplinkMessage = &v
		case "uplink_normalized", "uplinkNormalized":
			s.AddField("uplink_normalized")
			if s.ReadNil() {
				x.UplinkNormalized = nil
				return
			}
			// NOTE: ApplicationPubSub_Message does not seem to implement UnmarshalProtoJSON.
			var v ApplicationPubSub_Message
			golang.UnmarshalMessage(s, &v)
			x.UplinkNormalized = &v
		case "join_accept", "joinAccept":
			s.AddField("join_accept")
			if s.ReadNil() {
				x.JoinAccept = nil
				return
			}
			// NOTE: ApplicationPubSub_Message does not seem to implement UnmarshalProtoJSON.
			var v ApplicationPubSub_Message
			golang.UnmarshalMessage(s, &v)
			x.JoinAccept = &v
		case "downlink_ack", "downlinkAck":
			s.AddField("downlink_ack")
			if s.ReadNil() {
				x.DownlinkAck = nil
				return
			}
			// NOTE: ApplicationPubSub_Message does not seem to implement UnmarshalProtoJSON.
			var v ApplicationPubSub_Message
			golang.UnmarshalMessage(s, &v)
			x.DownlinkAck = &v
		case "downlink_nack", "downlinkNack":
			s.AddField("downlink_nack")
			if s.ReadNil() {
				x.DownlinkNack = nil
				return
			}
			// NOTE: ApplicationPubSub_Message does not seem to implement UnmarshalProtoJSON.
			var v ApplicationPubSub_Message
			golang.UnmarshalMessage(s, &v)
			x.DownlinkNack = &v
		case "downlink_sent", "downlinkSent":
			s.AddField("downlink_sent")
			if s.ReadNil() {
				x.DownlinkSent = nil
				return
			}
			// NOTE: ApplicationPubSub_Message does not seem to implement UnmarshalProtoJSON.
			var v ApplicationPubSub_Message
			golang.UnmarshalMessage(s, &v)
			x.DownlinkSent = &v
		case "downlink_failed", "downlinkFailed":
			s.AddField("downlink_failed")
			if s.ReadNil() {
				x.DownlinkFailed = nil
				return
			}
			// NOTE: ApplicationPubSub_Message does not seem to implement UnmarshalProtoJSON.
			var v ApplicationPubSub_Message
			golang.UnmarshalMessage(s, &v)
			x.DownlinkFailed = &v
		case "downlink_queued", "downlinkQueued":
			s.AddField("downlink_queued")
			if s.ReadNil() {
				x.DownlinkQueued = nil
				return
			}
			// NOTE: ApplicationPubSub_Message does not seem to implement UnmarshalProtoJSON.
			var v ApplicationPubSub_Message
			golang.UnmarshalMessage(s, &v)
			x.DownlinkQueued = &v
		case "downlink_queue_invalidated", "downlinkQueueInvalidated":
			s.AddField("downlink_queue_invalidated")
			if s.ReadNil() {
				x.DownlinkQueueInvalidated = nil
				return
			}
			// NOTE: ApplicationPubSub_Message does not seem to implement UnmarshalProtoJSON.
			var v ApplicationPubSub_Message
			golang.UnmarshalMessage(s, &v)
			x.DownlinkQueueInvalidated = &v
		case "location_solved", "locationSolved":
			s.AddField("location_solved")
			if s.ReadNil() {
				x.LocationSolved = nil
				return
			}
			// NOTE: ApplicationPubSub_Message does not seem to implement UnmarshalProtoJSON.
			var v ApplicationPubSub_Message
			golang.UnmarshalMessage(s, &v)
			x.LocationSolved = &v
		case "service_data", "serviceData":
			s.AddField("service_data")
			if s.ReadNil() {
				x.ServiceData = nil
				return
			}
			// NOTE: ApplicationPubSub_Message does not seem to implement UnmarshalProtoJSON.
			var v ApplicationPubSub_Message
			golang.UnmarshalMessage(s, &v)
			x.ServiceData = &v
		}
	})
}

// UnmarshalJSON unmarshals the ApplicationPubSub from JSON.
func (x *ApplicationPubSub) UnmarshalJSON(b []byte) error {
	return jsonplugin.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

// MarshalProtoJSON marshals the ApplicationPubSubs message to JSON.
func (x *ApplicationPubSubs) MarshalProtoJSON(s *jsonplugin.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if len(x.Pubsubs) > 0 || s.HasField("pubsubs") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("pubsubs")
		s.WriteArrayStart()
		var wroteElement bool
		for _, element := range x.Pubsubs {
			s.WriteMoreIf(&wroteElement)
			element.MarshalProtoJSON(s.WithField("pubsubs"))
		}
		s.WriteArrayEnd()
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the ApplicationPubSubs to JSON.
func (x *ApplicationPubSubs) MarshalJSON() ([]byte, error) {
	return jsonplugin.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the ApplicationPubSubs message from JSON.
func (x *ApplicationPubSubs) UnmarshalProtoJSON(s *jsonplugin.UnmarshalState) {
	if s.ReadNil() {
		return
	}
	s.ReadObject(func(key string) {
		switch key {
		default:
			s.ReadAny() // ignore unknown field
		case "pubsubs":
			s.AddField("pubsubs")
			if s.ReadNil() {
				x.Pubsubs = nil
				return
			}
			s.ReadArray(func() {
				if s.ReadNil() {
					x.Pubsubs = append(x.Pubsubs, nil)
					return
				}
				v := &ApplicationPubSub{}
				v.UnmarshalProtoJSON(s.WithField("pubsubs", false))
				if s.Err() != nil {
					return
				}
				x.Pubsubs = append(x.Pubsubs, v)
			})
		}
	})
}

// UnmarshalJSON unmarshals the ApplicationPubSubs from JSON.
func (x *ApplicationPubSubs) UnmarshalJSON(b []byte) error {
	return jsonplugin.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

// MarshalProtoJSON marshals the GetApplicationPubSubRequest message to JSON.
func (x *GetApplicationPubSubRequest) MarshalProtoJSON(s *jsonplugin.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if x.Ids != nil || s.HasField("ids") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("ids")
		// NOTE: ApplicationPubSubIdentifiers does not seem to implement MarshalProtoJSON.
		golang.MarshalMessage(s, x.Ids)
	}
	if x.FieldMask != nil || s.HasField("field_mask") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("field_mask")
		if x.FieldMask == nil {
			s.WriteNil()
		} else {
			golang.MarshalLegacyFieldMask(s, x.FieldMask)
		}
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the GetApplicationPubSubRequest to JSON.
func (x *GetApplicationPubSubRequest) MarshalJSON() ([]byte, error) {
	return jsonplugin.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the GetApplicationPubSubRequest message from JSON.
func (x *GetApplicationPubSubRequest) UnmarshalProtoJSON(s *jsonplugin.UnmarshalState) {
	if s.ReadNil() {
		return
	}
	s.ReadObject(func(key string) {
		switch key {
		default:
			s.ReadAny() // ignore unknown field
		case "ids":
			s.AddField("ids")
			if s.ReadNil() {
				x.Ids = nil
				return
			}
			// NOTE: ApplicationPubSubIdentifiers does not seem to implement UnmarshalProtoJSON.
			var v ApplicationPubSubIdentifiers
			golang.UnmarshalMessage(s, &v)
			x.Ids = &v
		case "field_mask", "fieldMask":
			s.AddField("field_mask")
			if s.ReadNil() {
				x.FieldMask = nil
				return
			}
			v := golang.UnmarshalFieldMask(s)
			if s.Err() != nil {
				return
			}
			x.FieldMask = v
		}
	})
}

// UnmarshalJSON unmarshals the GetApplicationPubSubRequest from JSON.
func (x *GetApplicationPubSubRequest) UnmarshalJSON(b []byte) error {
	return jsonplugin.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

// MarshalProtoJSON marshals the ListApplicationPubSubsRequest message to JSON.
func (x *ListApplicationPubSubsRequest) MarshalProtoJSON(s *jsonplugin.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if x.ApplicationIds != nil || s.HasField("application_ids") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("application_ids")
		// NOTE: ApplicationIdentifiers does not seem to implement MarshalProtoJSON.
		golang.MarshalMessage(s, x.ApplicationIds)
	}
	if x.FieldMask != nil || s.HasField("field_mask") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("field_mask")
		if x.FieldMask == nil {
			s.WriteNil()
		} else {
			golang.MarshalLegacyFieldMask(s, x.FieldMask)
		}
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the ListApplicationPubSubsRequest to JSON.
func (x *ListApplicationPubSubsRequest) MarshalJSON() ([]byte, error) {
	return jsonplugin.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the ListApplicationPubSubsRequest message from JSON.
func (x *ListApplicationPubSubsRequest) UnmarshalProtoJSON(s *jsonplugin.UnmarshalState) {
	if s.ReadNil() {
		return
	}
	s.ReadObject(func(key string) {
		switch key {
		default:
			s.ReadAny() // ignore unknown field
		case "application_ids", "applicationIds":
			s.AddField("application_ids")
			if s.ReadNil() {
				x.ApplicationIds = nil
				return
			}
			// NOTE: ApplicationIdentifiers does not seem to implement UnmarshalProtoJSON.
			var v ApplicationIdentifiers
			golang.UnmarshalMessage(s, &v)
			x.ApplicationIds = &v
		case "field_mask", "fieldMask":
			s.AddField("field_mask")
			if s.ReadNil() {
				x.FieldMask = nil
				return
			}
			v := golang.UnmarshalFieldMask(s)
			if s.Err() != nil {
				return
			}
			x.FieldMask = v
		}
	})
}

// UnmarshalJSON unmarshals the ListApplicationPubSubsRequest from JSON.
func (x *ListApplicationPubSubsRequest) UnmarshalJSON(b []byte) error {
	return jsonplugin.DefaultUnmarshalerConfig.Unmarshal(b, x)
}

// MarshalProtoJSON marshals the SetApplicationPubSubRequest message to JSON.
func (x *SetApplicationPubSubRequest) MarshalProtoJSON(s *jsonplugin.MarshalState) {
	if x == nil {
		s.WriteNil()
		return
	}
	s.WriteObjectStart()
	var wroteField bool
	if x.Pubsub != nil || s.HasField("pubsub") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("pubsub")
		x.Pubsub.MarshalProtoJSON(s.WithField("pubsub"))
	}
	if x.FieldMask != nil || s.HasField("field_mask") {
		s.WriteMoreIf(&wroteField)
		s.WriteObjectField("field_mask")
		if x.FieldMask == nil {
			s.WriteNil()
		} else {
			golang.MarshalLegacyFieldMask(s, x.FieldMask)
		}
	}
	s.WriteObjectEnd()
}

// MarshalJSON marshals the SetApplicationPubSubRequest to JSON.
func (x *SetApplicationPubSubRequest) MarshalJSON() ([]byte, error) {
	return jsonplugin.DefaultMarshalerConfig.Marshal(x)
}

// UnmarshalProtoJSON unmarshals the SetApplicationPubSubRequest message from JSON.
func (x *SetApplicationPubSubRequest) UnmarshalProtoJSON(s *jsonplugin.UnmarshalState) {
	if s.ReadNil() {
		return
	}
	s.ReadObject(func(key string) {
		switch key {
		default:
			s.ReadAny() // ignore unknown field
		case "pubsub":
			if s.ReadNil() {
				x.Pubsub = nil
				return
			}
			x.Pubsub = &ApplicationPubSub{}
			x.Pubsub.UnmarshalProtoJSON(s.WithField("pubsub", true))
		case "field_mask", "fieldMask":
			s.AddField("field_mask")
			if s.ReadNil() {
				x.FieldMask = nil
				return
			}
			v := golang.UnmarshalFieldMask(s)
			if s.Err() != nil {
				return
			}
			x.FieldMask = v
		}
	})
}

// UnmarshalJSON unmarshals the SetApplicationPubSubRequest from JSON.
func (x *SetApplicationPubSubRequest) UnmarshalJSON(b []byte) error {
	return jsonplugin.DefaultUnmarshalerConfig.Unmarshal(b, x)
}
