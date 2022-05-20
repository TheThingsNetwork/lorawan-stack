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
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// MessageHeader is the common LoRaWAN Backend Interfaces message header.
type MessageHeader struct {
	ProtocolVersion ProtocolVersion
	TransactionID   uint32
	MessageType     MessageType
	SenderID,
	ReceiverID string
	SenderNSID    *EUI64       `json:",omitempty"`
	ReceiverNSID  *EUI64       `json:",omitempty"`
	SenderToken   Buffer       `json:",omitempty"`
	ReceiverToken Buffer       `json:",omitempty"`
	VSExtension   *VSExtension `json:",omitempty"`
}

// AnswerHeader returns the header of the answer message.
func (h MessageHeader) AnswerHeader() (MessageHeader, error) {
	ansType, ok := h.MessageType.Answer()
	if !ok {
		return MessageHeader{}, errInvalidRequestType.WithAttributes("type", h.MessageType)
	}
	return MessageHeader{
		ProtocolVersion: h.ProtocolVersion,
		TransactionID:   h.TransactionID,
		MessageType:     ansType,
		ReceiverToken:   h.SenderToken,
		ReceiverID:      h.SenderID,
		ReceiverNSID:    h.SenderNSID,
		SenderID:        h.ReceiverID,
		SenderToken:     h.ReceiverToken,
		SenderNSID:      h.ReceiverNSID,
	}, nil
}

// VSExtension is a vendor-specific extension.
type VSExtension struct {
	VendorID VendorID
}

// Result contains the result of an operation.
type Result struct {
	ResultCode  ResultCode
	Description string `json:",omitempty"`
}

// ErrorMessage is a message with raw header and a result field.
type ErrorMessage struct {
	MessageHeader
	Result Result
}

// NsMessageHeader contains the message header for NS messages.
type NsMessageHeader struct {
	MessageHeader
	SenderID   NetID
	SenderNSID *EUI64 `json:",omitempty"`
}

// AsMessageHeader contains the message header for AS messages.
type AsMessageHeader struct {
	MessageHeader
}

// NsJsMessageHeader contains the message header for NS to JS messages.
type NsJsMessageHeader struct {
	MessageHeader
	SenderID   NetID
	SenderNSID *EUI64 `json:",omitempty"`
	// ReceiverID is a JoinEUI.
	ReceiverID EUI64
}

// JsNsMessageHeader contains the message header for JS to NS messages.
type JsNsMessageHeader struct {
	MessageHeader
	// SenderID is a JoinEUI.
	SenderID     EUI64
	ReceiverID   NetID
	ReceiverNSID *EUI64 `json:",omitempty"`
}

// AsJsMessageHeader contains the message header for AS to JS messages.
type AsJsMessageHeader struct {
	MessageHeader
	SenderID string
	// ReceiverID is a JoinEUI.
	ReceiverID EUI64
}

// JsAsMessageHeader contains the message header for JS to AS messages.
type JsAsMessageHeader struct {
	MessageHeader
	// SenderID is a JoinEUI.
	SenderID   EUI64
	ReceiverID string
}

// JoinReq is a join-request message.
type JoinReq struct {
	NsJsMessageHeader
	MACVersion MACVersion
	PHYPayload Buffer
	DevEUI     EUI64
	DevAddr    DevAddr
	DLSettings Buffer
	RxDelay    ttnpb.RxDelay
	CFList     Buffer
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

// AppSKeyReq is a AppSKey request message.
type AppSKeyReq struct {
	AsJsMessageHeader
	DevEUI       EUI64
	SessionKeyID Buffer
}

// AppSKeyAns is an answer to an AppSKeyReq message.
type AppSKeyAns struct {
	JsAsMessageHeader
	Result       Result
	DevEUI       EUI64
	AppSKey      *KeyEnvelope
	SessionKeyID Buffer
}

// HomeNSReq is a NetID request message.
type HomeNSReq struct {
	NsJsMessageHeader
	DevEUI EUI64
}

// HomeNSAns is an answer to a HomeNSReq message.
type HomeNSAns struct {
	JsNsMessageHeader
	Result Result
	HNSID  *EUI64 `json:",omitempty"`
	HNetID NetID
}
