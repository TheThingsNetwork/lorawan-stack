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

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.30.0
// 	protoc        v4.22.2
// source: lorawan-stack/api/join.proto

package ttnpb

import (
	_ "github.com/TheThingsIndustries/protoc-gen-go-json/annotations"
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	_ "github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2/options"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	durationpb "google.golang.org/protobuf/types/known/durationpb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type JoinRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RawPayload         []byte      `protobuf:"bytes,1,opt,name=raw_payload,json=rawPayload,proto3" json:"raw_payload,omitempty"`
	Payload            *Message    `protobuf:"bytes,2,opt,name=payload,proto3" json:"payload,omitempty"`
	DevAddr            []byte      `protobuf:"bytes,3,opt,name=dev_addr,json=devAddr,proto3" json:"dev_addr,omitempty"`
	SelectedMacVersion MACVersion  `protobuf:"varint,4,opt,name=selected_mac_version,json=selectedMacVersion,proto3,enum=ttn.lorawan.v3.MACVersion" json:"selected_mac_version,omitempty"`
	NetId              []byte      `protobuf:"bytes,5,opt,name=net_id,json=netId,proto3" json:"net_id,omitempty"`
	DownlinkSettings   *DLSettings `protobuf:"bytes,6,opt,name=downlink_settings,json=downlinkSettings,proto3" json:"downlink_settings,omitempty"`
	RxDelay            RxDelay     `protobuf:"varint,7,opt,name=rx_delay,json=rxDelay,proto3,enum=ttn.lorawan.v3.RxDelay" json:"rx_delay,omitempty"`
	// Optional CFList.
	CfList         *CFList  `protobuf:"bytes,8,opt,name=cf_list,json=cfList,proto3" json:"cf_list,omitempty"`
	CorrelationIds []string `protobuf:"bytes,10,rep,name=correlation_ids,json=correlationIds,proto3" json:"correlation_ids,omitempty"`
	// Consumed airtime for the transmission of the join request. Calculated by Network Server using the RawPayload size and the transmission settings.
	ConsumedAirtime *durationpb.Duration `protobuf:"bytes,11,opt,name=consumed_airtime,json=consumedAirtime,proto3" json:"consumed_airtime,omitempty"`
}

func (x *JoinRequest) Reset() {
	*x = JoinRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_lorawan_stack_api_join_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *JoinRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JoinRequest) ProtoMessage() {}

func (x *JoinRequest) ProtoReflect() protoreflect.Message {
	mi := &file_lorawan_stack_api_join_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JoinRequest.ProtoReflect.Descriptor instead.
func (*JoinRequest) Descriptor() ([]byte, []int) {
	return file_lorawan_stack_api_join_proto_rawDescGZIP(), []int{0}
}

func (x *JoinRequest) GetRawPayload() []byte {
	if x != nil {
		return x.RawPayload
	}
	return nil
}

func (x *JoinRequest) GetPayload() *Message {
	if x != nil {
		return x.Payload
	}
	return nil
}

func (x *JoinRequest) GetDevAddr() []byte {
	if x != nil {
		return x.DevAddr
	}
	return nil
}

func (x *JoinRequest) GetSelectedMacVersion() MACVersion {
	if x != nil {
		return x.SelectedMacVersion
	}
	return MACVersion_MAC_UNKNOWN
}

func (x *JoinRequest) GetNetId() []byte {
	if x != nil {
		return x.NetId
	}
	return nil
}

func (x *JoinRequest) GetDownlinkSettings() *DLSettings {
	if x != nil {
		return x.DownlinkSettings
	}
	return nil
}

func (x *JoinRequest) GetRxDelay() RxDelay {
	if x != nil {
		return x.RxDelay
	}
	return RxDelay_RX_DELAY_0
}

func (x *JoinRequest) GetCfList() *CFList {
	if x != nil {
		return x.CfList
	}
	return nil
}

func (x *JoinRequest) GetCorrelationIds() []string {
	if x != nil {
		return x.CorrelationIds
	}
	return nil
}

func (x *JoinRequest) GetConsumedAirtime() *durationpb.Duration {
	if x != nil {
		return x.ConsumedAirtime
	}
	return nil
}

type JoinResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RawPayload     []byte               `protobuf:"bytes,1,opt,name=raw_payload,json=rawPayload,proto3" json:"raw_payload,omitempty"`
	SessionKeys    *SessionKeys         `protobuf:"bytes,2,opt,name=session_keys,json=sessionKeys,proto3" json:"session_keys,omitempty"`
	Lifetime       *durationpb.Duration `protobuf:"bytes,3,opt,name=lifetime,proto3" json:"lifetime,omitempty"`
	CorrelationIds []string             `protobuf:"bytes,4,rep,name=correlation_ids,json=correlationIds,proto3" json:"correlation_ids,omitempty"`
}

func (x *JoinResponse) Reset() {
	*x = JoinResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_lorawan_stack_api_join_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *JoinResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*JoinResponse) ProtoMessage() {}

func (x *JoinResponse) ProtoReflect() protoreflect.Message {
	mi := &file_lorawan_stack_api_join_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use JoinResponse.ProtoReflect.Descriptor instead.
func (*JoinResponse) Descriptor() ([]byte, []int) {
	return file_lorawan_stack_api_join_proto_rawDescGZIP(), []int{1}
}

func (x *JoinResponse) GetRawPayload() []byte {
	if x != nil {
		return x.RawPayload
	}
	return nil
}

func (x *JoinResponse) GetSessionKeys() *SessionKeys {
	if x != nil {
		return x.SessionKeys
	}
	return nil
}

func (x *JoinResponse) GetLifetime() *durationpb.Duration {
	if x != nil {
		return x.Lifetime
	}
	return nil
}

func (x *JoinResponse) GetCorrelationIds() []string {
	if x != nil {
		return x.CorrelationIds
	}
	return nil
}

var File_lorawan_stack_api_join_proto protoreflect.FileDescriptor

var file_lorawan_stack_api_join_proto_rawDesc = []byte{
	0x0a, 0x1c, 0x6c, 0x6f, 0x72, 0x61, 0x77, 0x61, 0x6e, 0x2d, 0x73, 0x74, 0x61, 0x63, 0x6b, 0x2f,
	0x61, 0x70, 0x69, 0x2f, 0x6a, 0x6f, 0x69, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0e,
	0x74, 0x74, 0x6e, 0x2e, 0x6c, 0x6f, 0x72, 0x61, 0x77, 0x61, 0x6e, 0x2e, 0x76, 0x33, 0x1a, 0x41,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x65, 0x6e, 0x76, 0x6f, 0x79,
	0x70, 0x72, 0x6f, 0x78, 0x79, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e,
	0x2d, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x43, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x54, 0x68,
	0x65, 0x54, 0x68, 0x69, 0x6e, 0x67, 0x73, 0x49, 0x6e, 0x64, 0x75, 0x73, 0x74, 0x72, 0x69, 0x65,
	0x73, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x67, 0x6f, 0x2d,
	0x6a, 0x73, 0x6f, 0x6e, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x64, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x6c, 0x6f, 0x72, 0x61, 0x77, 0x61, 0x6e, 0x2d,
	0x73, 0x74, 0x61, 0x63, 0x6b, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x6b, 0x65, 0x79, 0x73, 0x2e, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x6c, 0x6f, 0x72, 0x61, 0x77, 0x61, 0x6e, 0x2d, 0x73, 0x74,
	0x61, 0x63, 0x6b, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x6c, 0x6f, 0x72, 0x61, 0x77, 0x61, 0x6e, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65,
	0x6e, 0x2d, 0x6f, 0x70, 0x65, 0x6e, 0x61, 0x70, 0x69, 0x76, 0x32, 0x2f, 0x6f, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x73, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x8d, 0x07, 0x0a, 0x0b, 0x4a, 0x6f, 0x69, 0x6e, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x28, 0x0a, 0x0b, 0x72, 0x61, 0x77, 0x5f, 0x70, 0x61, 0x79,
	0x6c, 0x6f, 0x61, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x42, 0x07, 0xfa, 0x42, 0x04, 0x7a,
	0x02, 0x68, 0x17, 0x52, 0x0a, 0x72, 0x61, 0x77, 0x50, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x12,
	0x31, 0x0a, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x17, 0x2e, 0x74, 0x74, 0x6e, 0x2e, 0x6c, 0x6f, 0x72, 0x61, 0x77, 0x61, 0x6e, 0x2e, 0x76,
	0x33, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x07, 0x70, 0x61, 0x79, 0x6c, 0x6f,
	0x61, 0x64, 0x12, 0xc8, 0x01, 0x0a, 0x08, 0x64, 0x65, 0x76, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x0c, 0x42, 0xac, 0x01, 0x92, 0x41, 0x19, 0x4a, 0x0a, 0x22, 0x32, 0x36,
	0x30, 0x30, 0x41, 0x42, 0x43, 0x44, 0x22, 0x9a, 0x02, 0x01, 0x07, 0xa2, 0x02, 0x06, 0x73, 0x74,
	0x72, 0x69, 0x6e, 0x67, 0xfa, 0x42, 0x06, 0x7a, 0x04, 0x68, 0x04, 0x70, 0x01, 0xea, 0xaa, 0x19,
	0x82, 0x01, 0x0a, 0x3f, 0x67, 0x6f, 0x2e, 0x74, 0x68, 0x65, 0x74, 0x68, 0x69, 0x6e, 0x67, 0x73,
	0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2f, 0x6c, 0x6f, 0x72, 0x61, 0x77, 0x61, 0x6e,
	0x2d, 0x73, 0x74, 0x61, 0x63, 0x6b, 0x2f, 0x76, 0x33, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x74, 0x79,
	0x70, 0x65, 0x73, 0x2e, 0x4d, 0x61, 0x72, 0x73, 0x68, 0x61, 0x6c, 0x48, 0x45, 0x58, 0x42, 0x79,
	0x74, 0x65, 0x73, 0x12, 0x3f, 0x67, 0x6f, 0x2e, 0x74, 0x68, 0x65, 0x74, 0x68, 0x69, 0x6e, 0x67,
	0x73, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2f, 0x6c, 0x6f, 0x72, 0x61, 0x77, 0x61,
	0x6e, 0x2d, 0x73, 0x74, 0x61, 0x63, 0x6b, 0x2f, 0x76, 0x33, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x74,
	0x79, 0x70, 0x65, 0x73, 0x2e, 0x55, 0x6e, 0x6d, 0x61, 0x72, 0x73, 0x68, 0x61, 0x6c, 0x34, 0x42,
	0x79, 0x74, 0x65, 0x73, 0x52, 0x07, 0x64, 0x65, 0x76, 0x41, 0x64, 0x64, 0x72, 0x12, 0x4c, 0x0a,
	0x14, 0x73, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x65, 0x64, 0x5f, 0x6d, 0x61, 0x63, 0x5f, 0x76, 0x65,
	0x72, 0x73, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1a, 0x2e, 0x74, 0x74,
	0x6e, 0x2e, 0x6c, 0x6f, 0x72, 0x61, 0x77, 0x61, 0x6e, 0x2e, 0x76, 0x33, 0x2e, 0x4d, 0x41, 0x43,
	0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x52, 0x12, 0x73, 0x65, 0x6c, 0x65, 0x63, 0x74, 0x65,
	0x64, 0x4d, 0x61, 0x63, 0x56, 0x65, 0x72, 0x73, 0x69, 0x6f, 0x6e, 0x12, 0xc2, 0x01, 0x0a, 0x06,
	0x6e, 0x65, 0x74, 0x5f, 0x69, 0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x42, 0xaa, 0x01, 0x92,
	0x41, 0x17, 0x4a, 0x08, 0x22, 0x30, 0x30, 0x30, 0x30, 0x31, 0x33, 0x22, 0x9a, 0x02, 0x01, 0x07,
	0xa2, 0x02, 0x06, 0x73, 0x74, 0x72, 0x69, 0x6e, 0x67, 0xfa, 0x42, 0x06, 0x7a, 0x04, 0x68, 0x03,
	0x70, 0x01, 0xea, 0xaa, 0x19, 0x82, 0x01, 0x0a, 0x3f, 0x67, 0x6f, 0x2e, 0x74, 0x68, 0x65, 0x74,
	0x68, 0x69, 0x6e, 0x67, 0x73, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2f, 0x6c, 0x6f,
	0x72, 0x61, 0x77, 0x61, 0x6e, 0x2d, 0x73, 0x74, 0x61, 0x63, 0x6b, 0x2f, 0x76, 0x33, 0x2f, 0x70,
	0x6b, 0x67, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x4d, 0x61, 0x72, 0x73, 0x68, 0x61, 0x6c,
	0x48, 0x45, 0x58, 0x42, 0x79, 0x74, 0x65, 0x73, 0x12, 0x3f, 0x67, 0x6f, 0x2e, 0x74, 0x68, 0x65,
	0x74, 0x68, 0x69, 0x6e, 0x67, 0x73, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x2f, 0x6c,
	0x6f, 0x72, 0x61, 0x77, 0x61, 0x6e, 0x2d, 0x73, 0x74, 0x61, 0x63, 0x6b, 0x2f, 0x76, 0x33, 0x2f,
	0x70, 0x6b, 0x67, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x55, 0x6e, 0x6d, 0x61, 0x72, 0x73,
	0x68, 0x61, 0x6c, 0x33, 0x42, 0x79, 0x74, 0x65, 0x73, 0x52, 0x05, 0x6e, 0x65, 0x74, 0x49, 0x64,
	0x12, 0x51, 0x0a, 0x11, 0x64, 0x6f, 0x77, 0x6e, 0x6c, 0x69, 0x6e, 0x6b, 0x5f, 0x73, 0x65, 0x74,
	0x74, 0x69, 0x6e, 0x67, 0x73, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x74, 0x74,
	0x6e, 0x2e, 0x6c, 0x6f, 0x72, 0x61, 0x77, 0x61, 0x6e, 0x2e, 0x76, 0x33, 0x2e, 0x44, 0x4c, 0x53,
	0x65, 0x74, 0x74, 0x69, 0x6e, 0x67, 0x73, 0x42, 0x08, 0xfa, 0x42, 0x05, 0x8a, 0x01, 0x02, 0x10,
	0x01, 0x52, 0x10, 0x64, 0x6f, 0x77, 0x6e, 0x6c, 0x69, 0x6e, 0x6b, 0x53, 0x65, 0x74, 0x74, 0x69,
	0x6e, 0x67, 0x73, 0x12, 0x3c, 0x0a, 0x08, 0x72, 0x78, 0x5f, 0x64, 0x65, 0x6c, 0x61, 0x79, 0x18,
	0x07, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x17, 0x2e, 0x74, 0x74, 0x6e, 0x2e, 0x6c, 0x6f, 0x72, 0x61,
	0x77, 0x61, 0x6e, 0x2e, 0x76, 0x33, 0x2e, 0x52, 0x78, 0x44, 0x65, 0x6c, 0x61, 0x79, 0x42, 0x08,
	0xfa, 0x42, 0x05, 0x82, 0x01, 0x02, 0x10, 0x01, 0x52, 0x07, 0x72, 0x78, 0x44, 0x65, 0x6c, 0x61,
	0x79, 0x12, 0x2f, 0x0a, 0x07, 0x63, 0x66, 0x5f, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x08, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x16, 0x2e, 0x74, 0x74, 0x6e, 0x2e, 0x6c, 0x6f, 0x72, 0x61, 0x77, 0x61, 0x6e,
	0x2e, 0x76, 0x33, 0x2e, 0x43, 0x46, 0x4c, 0x69, 0x73, 0x74, 0x52, 0x06, 0x63, 0x66, 0x4c, 0x69,
	0x73, 0x74, 0x12, 0x35, 0x0a, 0x0f, 0x63, 0x6f, 0x72, 0x72, 0x65, 0x6c, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x0a, 0x20, 0x03, 0x28, 0x09, 0x42, 0x0c, 0xfa, 0x42, 0x09,
	0x92, 0x01, 0x06, 0x22, 0x04, 0x72, 0x02, 0x18, 0x64, 0x52, 0x0e, 0x63, 0x6f, 0x72, 0x72, 0x65,
	0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x73, 0x12, 0x44, 0x0a, 0x10, 0x63, 0x6f, 0x6e,
	0x73, 0x75, 0x6d, 0x65, 0x64, 0x5f, 0x61, 0x69, 0x72, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x0b, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x19, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x0f,
	0x63, 0x6f, 0x6e, 0x73, 0x75, 0x6d, 0x65, 0x64, 0x41, 0x69, 0x72, 0x74, 0x69, 0x6d, 0x65, 0x4a,
	0x04, 0x08, 0x09, 0x10, 0x0a, 0x22, 0xf2, 0x01, 0x0a, 0x0c, 0x4a, 0x6f, 0x69, 0x6e, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x2a, 0x0a, 0x0b, 0x72, 0x61, 0x77, 0x5f, 0x70, 0x61,
	0x79, 0x6c, 0x6f, 0x61, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x42, 0x09, 0xfa, 0x42, 0x06,
	0x7a, 0x04, 0x10, 0x11, 0x18, 0x21, 0x52, 0x0a, 0x72, 0x61, 0x77, 0x50, 0x61, 0x79, 0x6c, 0x6f,
	0x61, 0x64, 0x12, 0x48, 0x0a, 0x0c, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x5f, 0x6b, 0x65,
	0x79, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x1b, 0x2e, 0x74, 0x74, 0x6e, 0x2e, 0x6c,
	0x6f, 0x72, 0x61, 0x77, 0x61, 0x6e, 0x2e, 0x76, 0x33, 0x2e, 0x53, 0x65, 0x73, 0x73, 0x69, 0x6f,
	0x6e, 0x4b, 0x65, 0x79, 0x73, 0x42, 0x08, 0xfa, 0x42, 0x05, 0x8a, 0x01, 0x02, 0x10, 0x01, 0x52,
	0x0b, 0x73, 0x65, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x4b, 0x65, 0x79, 0x73, 0x12, 0x35, 0x0a, 0x08,
	0x6c, 0x69, 0x66, 0x65, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x19,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x44, 0x75, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x52, 0x08, 0x6c, 0x69, 0x66, 0x65, 0x74,
	0x69, 0x6d, 0x65, 0x12, 0x35, 0x0a, 0x0f, 0x63, 0x6f, 0x72, 0x72, 0x65, 0x6c, 0x61, 0x74, 0x69,
	0x6f, 0x6e, 0x5f, 0x69, 0x64, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x09, 0x42, 0x0c, 0xfa, 0x42,
	0x09, 0x92, 0x01, 0x06, 0x22, 0x04, 0x72, 0x02, 0x18, 0x64, 0x52, 0x0e, 0x63, 0x6f, 0x72, 0x72,
	0x65, 0x6c, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x49, 0x64, 0x73, 0x42, 0x31, 0x5a, 0x2f, 0x67, 0x6f,
	0x2e, 0x74, 0x68, 0x65, 0x74, 0x68, 0x69, 0x6e, 0x67, 0x73, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f,
	0x72, 0x6b, 0x2f, 0x6c, 0x6f, 0x72, 0x61, 0x77, 0x61, 0x6e, 0x2d, 0x73, 0x74, 0x61, 0x63, 0x6b,
	0x2f, 0x76, 0x33, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x74, 0x74, 0x6e, 0x70, 0x62, 0x62, 0x06, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_lorawan_stack_api_join_proto_rawDescOnce sync.Once
	file_lorawan_stack_api_join_proto_rawDescData = file_lorawan_stack_api_join_proto_rawDesc
)

func file_lorawan_stack_api_join_proto_rawDescGZIP() []byte {
	file_lorawan_stack_api_join_proto_rawDescOnce.Do(func() {
		file_lorawan_stack_api_join_proto_rawDescData = protoimpl.X.CompressGZIP(file_lorawan_stack_api_join_proto_rawDescData)
	})
	return file_lorawan_stack_api_join_proto_rawDescData
}

var file_lorawan_stack_api_join_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_lorawan_stack_api_join_proto_goTypes = []interface{}{
	(*JoinRequest)(nil),         // 0: ttn.lorawan.v3.JoinRequest
	(*JoinResponse)(nil),        // 1: ttn.lorawan.v3.JoinResponse
	(*Message)(nil),             // 2: ttn.lorawan.v3.Message
	(MACVersion)(0),             // 3: ttn.lorawan.v3.MACVersion
	(*DLSettings)(nil),          // 4: ttn.lorawan.v3.DLSettings
	(RxDelay)(0),                // 5: ttn.lorawan.v3.RxDelay
	(*CFList)(nil),              // 6: ttn.lorawan.v3.CFList
	(*durationpb.Duration)(nil), // 7: google.protobuf.Duration
	(*SessionKeys)(nil),         // 8: ttn.lorawan.v3.SessionKeys
}
var file_lorawan_stack_api_join_proto_depIdxs = []int32{
	2, // 0: ttn.lorawan.v3.JoinRequest.payload:type_name -> ttn.lorawan.v3.Message
	3, // 1: ttn.lorawan.v3.JoinRequest.selected_mac_version:type_name -> ttn.lorawan.v3.MACVersion
	4, // 2: ttn.lorawan.v3.JoinRequest.downlink_settings:type_name -> ttn.lorawan.v3.DLSettings
	5, // 3: ttn.lorawan.v3.JoinRequest.rx_delay:type_name -> ttn.lorawan.v3.RxDelay
	6, // 4: ttn.lorawan.v3.JoinRequest.cf_list:type_name -> ttn.lorawan.v3.CFList
	7, // 5: ttn.lorawan.v3.JoinRequest.consumed_airtime:type_name -> google.protobuf.Duration
	8, // 6: ttn.lorawan.v3.JoinResponse.session_keys:type_name -> ttn.lorawan.v3.SessionKeys
	7, // 7: ttn.lorawan.v3.JoinResponse.lifetime:type_name -> google.protobuf.Duration
	8, // [8:8] is the sub-list for method output_type
	8, // [8:8] is the sub-list for method input_type
	8, // [8:8] is the sub-list for extension type_name
	8, // [8:8] is the sub-list for extension extendee
	0, // [0:8] is the sub-list for field type_name
}

func init() { file_lorawan_stack_api_join_proto_init() }
func file_lorawan_stack_api_join_proto_init() {
	if File_lorawan_stack_api_join_proto != nil {
		return
	}
	file_lorawan_stack_api_keys_proto_init()
	file_lorawan_stack_api_lorawan_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_lorawan_stack_api_join_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*JoinRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_lorawan_stack_api_join_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*JoinResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_lorawan_stack_api_join_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_lorawan_stack_api_join_proto_goTypes,
		DependencyIndexes: file_lorawan_stack_api_join_proto_depIdxs,
		MessageInfos:      file_lorawan_stack_api_join_proto_msgTypes,
	}.Build()
	File_lorawan_stack_api_join_proto = out.File
	file_lorawan_stack_api_join_proto_rawDesc = nil
	file_lorawan_stack_api_join_proto_goTypes = nil
	file_lorawan_stack_api_join_proto_depIdxs = nil
}
