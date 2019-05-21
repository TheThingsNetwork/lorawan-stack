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

package interop

import (
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"net/http"

	echo "github.com/labstack/echo/v4"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// MessageHeader contains the message header.
type MessageHeader struct {
	ProtocolVersion string
	TransactionID   uint32
	MessageType     MessageType
	SenderToken     Buffer `json:",omitempty"`
	ReceiverToken   Buffer `json:",omitempty"`
}

// AnswerHeader returns the header of the answer message.
func (h MessageHeader) AnswerHeader() (MessageHeader, error) {
	var ansType MessageType
	switch h.MessageType {
	case MessageTypeJoinReq:
		ansType = MessageTypeJoinAns
	case MessageTypeRejoinReq:
		ansType = MessageTypeRejoinAns
	case MessageTypeAppSKeyReq:
		ansType = MessageTypeAppSKeyAns
	case MessageTypePRStartReq:
		ansType = MessageTypePRStartAns
	case MessageTypePRStopReq:
		ansType = MessageTypePRStopAns
	case MessageTypeHRStartReq:
		ansType = MessageTypeHRStartAns
	case MessageTypeHRStopReq:
		ansType = MessageTypeHRStopAns
	case MessageTypeHomeNSReq:
		ansType = MessageTypeHomeNSAns
	case MessageTypeProfileReq:
		ansType = MessageTypeProfileAns
	case MessageTypeXmitDataReq:
		ansType = MessageTypeXmitDataAns
	case MessageTypeXmitLocReq:
		ansType = MessageTypeXmitLocAns
	default:
		return MessageHeader{}, errInvalidRequestType.WithAttributes("type", h.MessageType)
	}
	return MessageHeader{
		ProtocolVersion: h.ProtocolVersion,
		TransactionID:   h.TransactionID,
		MessageType:     ansType,
		ReceiverToken:   h.SenderToken,
	}, nil
}

// RawMessageHeader contains a message header with generic sender and receiver IDs.
type RawMessageHeader struct {
	MessageHeader
	SenderID,
	ReceiverID string
}

// AnswerHeader returns the header of the answer message.
func (h RawMessageHeader) AnswerHeader() (RawMessageHeader, error) {
	header, err := h.MessageHeader.AnswerHeader()
	if err != nil {
		return RawMessageHeader{}, err
	}
	return RawMessageHeader{
		MessageHeader: header,
		SenderID:      h.ReceiverID,
		ReceiverID:    h.SenderID,
	}, nil
}

// ErrorMessage is a message with raw header and a result field.
type ErrorMessage struct {
	RawMessageHeader
	Result Result
}

// NsJsMessageHeader contains the message header for NS to JS messages.
type NsJsMessageHeader struct {
	MessageHeader
	SenderID types.NetID
	// ReceiverID is a JoinEUI.
	ReceiverID types.EUI64
}

// AnswerHeader returns the header of the answer message.
func (h NsJsMessageHeader) AnswerHeader() (JsNsMessageHeader, error) {
	header, err := h.MessageHeader.AnswerHeader()
	if err != nil {
		return JsNsMessageHeader{}, err
	}
	return JsNsMessageHeader{
		MessageHeader: header,
		SenderID:      h.ReceiverID,
		ReceiverID:    h.SenderID,
	}, nil
}

// JsNsMessageHeader contains the message header for JS to NS messages.
type JsNsMessageHeader struct {
	MessageHeader
	// SenderID is a JoinEUI.
	SenderID   types.EUI64
	ReceiverID types.NetID
}

// JoinReq is a join-request message.
type JoinReq struct {
	NsJsMessageHeader
	MACVersion MACVersion
	PHYPayload Buffer
	DevEUI     types.EUI64
	DevAddr    types.DevAddr
	DLSettings Buffer
	RxDelay    ttnpb.RxDelay
	CFList     Buffer
	CFListType ttnpb.CFListType
}

// JoinAns is an answer to a JoinReq message.
type JoinAns struct {
	JsNsMessageHeader
	PHYPayload   Buffer
	Result       Result
	Lifetime     uint32
	SNwkSIntKey  *KeyEnvelope `json:",omitempty"`
	FNwkSIntKey  *KeyEnvelope `json:",omitempty"`
	NwkSEncKey   *KeyEnvelope `json:",omitempty"`
	NwkSKey      *KeyEnvelope `json:",omitempty"`
	AppSKey      *KeyEnvelope `json:",omitempty"`
	SessionKeyID Buffer       `json:",omitempty"`
}

// parseMessage parses the header and the message type of the request body.
// This middleware sets the header in the context on the `headerKey` and the message on the `messageKey`.
func parseMessage() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			buf, err := ioutil.ReadAll(c.Request().Body)
			if err != nil {
				return err
			}
			if len(buf) == 0 {
				c.NoContent(http.StatusBadRequest)
				return nil
			}
			var header RawMessageHeader
			if err := json.Unmarshal(buf, &header); err != nil {
				c.NoContent(http.StatusBadRequest)
				return nil
			}
			if header.ProtocolVersion == "" || header.MessageType == "" {
				c.NoContent(http.StatusBadRequest)
				return nil
			}
			c.Set(headerKey, &header)
			switch header.ProtocolVersion {
			case "1.0", "1.1":
			default:
				return ErrProtocolVersion
			}
			var msg interface{}
			switch header.MessageType {
			case MessageTypeJoinReq:
				msg = &JoinReq{}
			default:
				return ErrMalformedMessage
			}
			if err := json.Unmarshal(buf, msg); err != nil {
				return ErrMalformedMessage
			}
			c.Set(messageKey, msg)
			return next(c)
		}
	}
}

// verifySenderID verifies whether the SenderID of the message is authorized for the request according to the trusted
// certificates that are provided through the given callback.
func verifySenderID(getSenderClientCAs func(string) []*x509.Certificate) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Get(headerKey).(*RawMessageHeader)
			senderClientCAs := getSenderClientCAs(header.SenderID)
			if len(senderClientCAs) == 0 {
				c.NoContent(http.StatusForbidden)
				return nil
			}
			if state := c.Request().TLS; state != nil {
				for _, chain := range state.VerifiedChains {
					for _, cert := range chain {
						for _, senderCA := range senderClientCAs {
							if cert.Equal(senderCA) {
								return next(c)
							}
						}
					}
				}
			}
			// TODO: Check headers (https://github.com/TheThingsNetwork/lorawan-stack/issues/717)
			c.NoContent(http.StatusForbidden)
			return nil
		}
	}
}
