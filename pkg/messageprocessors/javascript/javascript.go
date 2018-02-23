// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

// Package javascript contains the Javascript payload formatter message processors.
package javascript

import (
	"context"
	"fmt"
	"reflect"

	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/gogoproto"
	"github.com/TheThingsNetwork/ttn/pkg/messageprocessors"
	"github.com/TheThingsNetwork/ttn/pkg/scripting"
	js "github.com/TheThingsNetwork/ttn/pkg/scripting/javascript"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
)

type host struct {
	ctx    context.Context
	engine scripting.Engine
}

// New creates and returns a new Javascript payload encoder and decoder.
func New(ctx context.Context) messageprocessors.PayloadEncodeDecoder {
	return &host{
		ctx:    ctx,
		engine: js.New(scripting.DefaultOptions),
	}
}

func (h *host) createEnvironment(model *ttnpb.EndDeviceModel) map[string]interface{} {
	env := make(map[string]interface{})
	env["brand"] = model.Brand
	env["model"] = model.Model
	env["hardware_version"] = model.HardwareVersion
	env["firmware_version"] = model.FirmwareVersion
	return env
}

// Encode encodes the message's MAC payload DecodedPayload to FRMPayload using script.
func (h *host) Encode(model *ttnpb.EndDeviceModel, script string, msg *ttnpb.DownlinkMessage) (*ttnpb.DownlinkMessage, error) {
	payload := msg.Payload.GetMACPayload()
	if payload == nil {
		return nil, messageprocessors.ErrNotMACPayload.New(nil)
	}

	decoded := payload.DecodedPayload
	if decoded == nil {
		return msg, nil
	}

	m, err := gogoproto.Map(decoded)
	if err != nil {
		return nil, scripting.ErrInvalidInputType.NewWithCause(nil, err)
	}

	env := h.createEnvironment(model)
	env["application_id"] = msg.ApplicationID
	env["device_id"] = msg.DeviceID
	env["deveui"] = msg.DevEUI
	env["joineui"] = msg.JoinEUI
	env["payload"] = m
	env["fport"] = payload.FPort

	script = fmt.Sprintf(`
		%s
		Encoder(env.payload, env.fport)
	`, script)

	value, err := h.engine.Run(h.ctx, script, env)
	if err != nil {
		return nil, err
	}

	slice := reflect.ValueOf(value)
	l := slice.Len()
	payload.FRMPayload = make([]byte, l)
	for i := 0; i < l; i++ {
		val := slice.Index(i).Interface()
		var b int64
		switch i := val.(type) {
		case int:
			b = int64(i)
		case int8:
			b = int64(i)
		case int16:
			b = int64(i)
		case int32:
			b = int64(i)
		case int64:
			b = int64(i)
		case uint8:
			b = int64(i)
		case uint16:
			b = int64(i)
		case uint32:
			b = int64(i)
		case uint64:
			b = i
		default:
			return nil, scripting.ErrInvalidOutputType.New(errors.Attributes{
				"type": fmt.Sprintf("%T[]", val),
			})
		}
		if b < 0x00 || b > 0xFF {
			return nil, scripting.ErrInvalidOutputRange.New(errors.Attributes{
				"value": b,
				"low":   0x00,
				"high":  0xFF,
			})
		}
		payload.FRMPayload[i] = byte(b)
	}

	return msg, nil
}

// Decode decodes the message's MAC payload FRMPayload to DecodedPayload using script.
func (h *host) Decode(model *ttnpb.EndDeviceModel, script string, msg *ttnpb.UplinkMessage) (*ttnpb.UplinkMessage, error) {
	payload := msg.Payload.GetMACPayload()
	if payload == nil {
		return nil, messageprocessors.ErrNotMACPayload.New(nil)
	}

	env := h.createEnvironment(model)
	env["application_id"] = msg.ApplicationID
	env["device_id"] = msg.DeviceID
	env["deveui"] = msg.DevEUI
	env["joineui"] = msg.JoinEUI
	env["payload"] = payload.FRMPayload
	env["fport"] = payload.FPort

	script = fmt.Sprintf(`
		%s
		Decoder(env.payload, env.fport)
	`, script)

	value, err := h.engine.Run(h.ctx, script, env)
	if err != nil {
		return nil, err
	}

	m, ok := value.(map[string]interface{})
	if !ok {
		return nil, scripting.ErrInvalidOutputType.New(errors.Attributes{
			"type": fmt.Sprintf("%T", value),
		})
	}

	s, err := gogoproto.Struct(m)
	if err != nil {
		return nil, scripting.ErrInvalidOutputType.NewWithCause(errors.Attributes{
			"type": fmt.Sprintf("%#v", value),
		}, err)
	}

	payload.DecodedPayload = s
	return msg, nil
}
