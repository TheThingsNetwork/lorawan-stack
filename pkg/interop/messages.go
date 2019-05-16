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
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
)

// MessageHeader contains the message header.
type MessageHeader struct {
	ProtocolVersion string
	TransactionID   uint32
	MessageType     MessageType
	SenderToken     Buffer
	ReceiverToken   Buffer
}

// NsJsMessageHeader contains the message header for NS to JS messages.
type NsJsMessageHeader struct {
	MessageHeader
	SenderID types.NetID
	// ReceiverID is a JoinEUI.
	ReceiverID types.EUI64
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
