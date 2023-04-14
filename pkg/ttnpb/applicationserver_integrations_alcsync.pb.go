// Copyright © 2023 The Things Network Foundation, The Things Industries B.V.
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
// source: lorawan-stack/api/applicationserver_integrations_alcsync.proto

package ttnpb

import (
	_ "github.com/TheThingsIndustries/protoc-gen-go-json/annotations"
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ALCSyncCommandIdentifier int32

const (
	ALCSyncCommandIdentifier_ALCSYNC_CID_PKG_VERSION              ALCSyncCommandIdentifier = 0
	ALCSyncCommandIdentifier_ALCSYNC_CID_APP_TIME                 ALCSyncCommandIdentifier = 1
	ALCSyncCommandIdentifier_ALCSYNC_CID_APP_DEV_TIME_PERIODICITY ALCSyncCommandIdentifier = 2
	ALCSyncCommandIdentifier_ALCSYNC_CID_FORCE_DEV_RESYNC         ALCSyncCommandIdentifier = 3
)

// Enum value maps for ALCSyncCommandIdentifier.
var (
	ALCSyncCommandIdentifier_name = map[int32]string{
		0: "ALCSYNC_CID_PKG_VERSION",
		1: "ALCSYNC_CID_APP_TIME",
		2: "ALCSYNC_CID_APP_DEV_TIME_PERIODICITY",
		3: "ALCSYNC_CID_FORCE_DEV_RESYNC",
	}
	ALCSyncCommandIdentifier_value = map[string]int32{
		"ALCSYNC_CID_PKG_VERSION":              0,
		"ALCSYNC_CID_APP_TIME":                 1,
		"ALCSYNC_CID_APP_DEV_TIME_PERIODICITY": 2,
		"ALCSYNC_CID_FORCE_DEV_RESYNC":         3,
	}
)

func (x ALCSyncCommandIdentifier) Enum() *ALCSyncCommandIdentifier {
	p := new(ALCSyncCommandIdentifier)
	*p = x
	return p
}

func (x ALCSyncCommandIdentifier) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (ALCSyncCommandIdentifier) Descriptor() protoreflect.EnumDescriptor {
	return file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_enumTypes[0].Descriptor()
}

func (ALCSyncCommandIdentifier) Type() protoreflect.EnumType {
	return &file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_enumTypes[0]
}

func (x ALCSyncCommandIdentifier) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use ALCSyncCommandIdentifier.Descriptor instead.
func (ALCSyncCommandIdentifier) EnumDescriptor() ([]byte, []int) {
	return file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_rawDescGZIP(), []int{0}
}

type ALCSyncCommand struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Cid ALCSyncCommandIdentifier `protobuf:"varint,1,opt,name=cid,proto3,enum=ttn.lorawan.v3.ALCSyncCommandIdentifier" json:"cid,omitempty"`
	// Types that are assignable to Payload:
	//	*ALCSyncCommand_AppTimeReq_
	//	*ALCSyncCommand_AppTimeAns_
	Payload isALCSyncCommand_Payload `protobuf_oneof:"payload"`
}

func (x *ALCSyncCommand) Reset() {
	*x = ALCSyncCommand{}
	if protoimpl.UnsafeEnabled {
		mi := &file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ALCSyncCommand) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ALCSyncCommand) ProtoMessage() {}

func (x *ALCSyncCommand) ProtoReflect() protoreflect.Message {
	mi := &file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ALCSyncCommand.ProtoReflect.Descriptor instead.
func (*ALCSyncCommand) Descriptor() ([]byte, []int) {
	return file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_rawDescGZIP(), []int{0}
}

func (x *ALCSyncCommand) GetCid() ALCSyncCommandIdentifier {
	if x != nil {
		return x.Cid
	}
	return ALCSyncCommandIdentifier_ALCSYNC_CID_PKG_VERSION
}

func (m *ALCSyncCommand) GetPayload() isALCSyncCommand_Payload {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (x *ALCSyncCommand) GetAppTimeReq() *ALCSyncCommand_AppTimeReq {
	if x, ok := x.GetPayload().(*ALCSyncCommand_AppTimeReq_); ok {
		return x.AppTimeReq
	}
	return nil
}

func (x *ALCSyncCommand) GetAppTimeAns() *ALCSyncCommand_AppTimeAns {
	if x, ok := x.GetPayload().(*ALCSyncCommand_AppTimeAns_); ok {
		return x.AppTimeAns
	}
	return nil
}

type isALCSyncCommand_Payload interface {
	isALCSyncCommand_Payload()
}

type ALCSyncCommand_AppTimeReq_ struct {
	AppTimeReq *ALCSyncCommand_AppTimeReq `protobuf:"bytes,2,opt,name=app_time_req,json=appTimeReq,proto3,oneof"`
}

type ALCSyncCommand_AppTimeAns_ struct {
	AppTimeAns *ALCSyncCommand_AppTimeAns `protobuf:"bytes,3,opt,name=app_time_ans,json=appTimeAns,proto3,oneof"`
}

func (*ALCSyncCommand_AppTimeReq_) isALCSyncCommand_Payload() {}

func (*ALCSyncCommand_AppTimeAns_) isALCSyncCommand_Payload() {}

type ALCSyncCommand_AppTimeReq struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	DeviceTime  *timestamppb.Timestamp `protobuf:"bytes,1,opt,name=DeviceTime,proto3" json:"DeviceTime,omitempty"`
	TokenReq    uint32                 `protobuf:"varint,2,opt,name=TokenReq,proto3" json:"TokenReq,omitempty"`
	AnsRequired bool                   `protobuf:"varint,3,opt,name=AnsRequired,proto3" json:"AnsRequired,omitempty"`
}

func (x *ALCSyncCommand_AppTimeReq) Reset() {
	*x = ALCSyncCommand_AppTimeReq{}
	if protoimpl.UnsafeEnabled {
		mi := &file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ALCSyncCommand_AppTimeReq) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ALCSyncCommand_AppTimeReq) ProtoMessage() {}

func (x *ALCSyncCommand_AppTimeReq) ProtoReflect() protoreflect.Message {
	mi := &file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ALCSyncCommand_AppTimeReq.ProtoReflect.Descriptor instead.
func (*ALCSyncCommand_AppTimeReq) Descriptor() ([]byte, []int) {
	return file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_rawDescGZIP(), []int{0, 0}
}

func (x *ALCSyncCommand_AppTimeReq) GetDeviceTime() *timestamppb.Timestamp {
	if x != nil {
		return x.DeviceTime
	}
	return nil
}

func (x *ALCSyncCommand_AppTimeReq) GetTokenReq() uint32 {
	if x != nil {
		return x.TokenReq
	}
	return 0
}

func (x *ALCSyncCommand_AppTimeReq) GetAnsRequired() bool {
	if x != nil {
		return x.AnsRequired
	}
	return false
}

type ALCSyncCommand_AppTimeAns struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	TimeCorrection int32  `protobuf:"varint,1,opt,name=TimeCorrection,proto3" json:"TimeCorrection,omitempty"`
	TokenAns       uint32 `protobuf:"varint,2,opt,name=TokenAns,proto3" json:"TokenAns,omitempty"`
}

func (x *ALCSyncCommand_AppTimeAns) Reset() {
	*x = ALCSyncCommand_AppTimeAns{}
	if protoimpl.UnsafeEnabled {
		mi := &file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ALCSyncCommand_AppTimeAns) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ALCSyncCommand_AppTimeAns) ProtoMessage() {}

func (x *ALCSyncCommand_AppTimeAns) ProtoReflect() protoreflect.Message {
	mi := &file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ALCSyncCommand_AppTimeAns.ProtoReflect.Descriptor instead.
func (*ALCSyncCommand_AppTimeAns) Descriptor() ([]byte, []int) {
	return file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_rawDescGZIP(), []int{0, 1}
}

func (x *ALCSyncCommand_AppTimeAns) GetTimeCorrection() int32 {
	if x != nil {
		return x.TimeCorrection
	}
	return 0
}

func (x *ALCSyncCommand_AppTimeAns) GetTokenAns() uint32 {
	if x != nil {
		return x.TokenAns
	}
	return 0
}

var File_lorawan_stack_api_applicationserver_integrations_alcsync_proto protoreflect.FileDescriptor

var file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_rawDesc = []byte{
	0x0a, 0x3e, 0x6c, 0x6f, 0x72, 0x61, 0x77, 0x61, 0x6e, 0x2d, 0x73, 0x74, 0x61, 0x63, 0x6b, 0x2f,
	0x61, 0x70, 0x69, 0x2f, 0x61, 0x70, 0x70, 0x6c, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x65, 0x72, 0x76, 0x65, 0x72, 0x5f, 0x69, 0x6e, 0x74, 0x65, 0x67, 0x72, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x5f, 0x61, 0x6c, 0x63, 0x73, 0x79, 0x6e, 0x63, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x0e, 0x74, 0x74, 0x6e, 0x2e, 0x6c, 0x6f, 0x72, 0x61, 0x77, 0x61, 0x6e, 0x2e, 0x76, 0x33,
	0x1a, 0x43, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x54, 0x68, 0x65,
	0x54, 0x68, 0x69, 0x6e, 0x67, 0x73, 0x49, 0x6e, 0x64, 0x75, 0x73, 0x74, 0x72, 0x69, 0x65, 0x73,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x67, 0x6f, 0x2d, 0x6a,
	0x73, 0x6f, 0x6e, 0x2f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x41, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x65, 0x6e, 0x76, 0x6f, 0x79, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x2f, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x63, 0x2d, 0x67, 0x65, 0x6e, 0x2d, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65,
	0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61, 0x74, 0x65, 0x2f, 0x76, 0x61, 0x6c, 0x69, 0x64, 0x61,
	0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1f, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69, 0x6d, 0x65, 0x73, 0x74,
	0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xf8, 0x03, 0x0a, 0x0e, 0x41, 0x4c,
	0x43, 0x53, 0x79, 0x6e, 0x63, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x12, 0x44, 0x0a, 0x03,
	0x63, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x28, 0x2e, 0x74, 0x74, 0x6e, 0x2e,
	0x6c, 0x6f, 0x72, 0x61, 0x77, 0x61, 0x6e, 0x2e, 0x76, 0x33, 0x2e, 0x41, 0x4c, 0x43, 0x53, 0x79,
	0x6e, 0x63, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66,
	0x69, 0x65, 0x72, 0x42, 0x08, 0xfa, 0x42, 0x05, 0x82, 0x01, 0x02, 0x10, 0x01, 0x52, 0x03, 0x63,
	0x69, 0x64, 0x12, 0x4d, 0x0a, 0x0c, 0x61, 0x70, 0x70, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x5f, 0x72,
	0x65, 0x71, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x74, 0x74, 0x6e, 0x2e, 0x6c,
	0x6f, 0x72, 0x61, 0x77, 0x61, 0x6e, 0x2e, 0x76, 0x33, 0x2e, 0x41, 0x4c, 0x43, 0x53, 0x79, 0x6e,
	0x63, 0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x2e, 0x41, 0x70, 0x70, 0x54, 0x69, 0x6d, 0x65,
	0x52, 0x65, 0x71, 0x48, 0x00, 0x52, 0x0a, 0x61, 0x70, 0x70, 0x54, 0x69, 0x6d, 0x65, 0x52, 0x65,
	0x71, 0x12, 0x4d, 0x0a, 0x0c, 0x61, 0x70, 0x70, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x5f, 0x61, 0x6e,
	0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x29, 0x2e, 0x74, 0x74, 0x6e, 0x2e, 0x6c, 0x6f,
	0x72, 0x61, 0x77, 0x61, 0x6e, 0x2e, 0x76, 0x33, 0x2e, 0x41, 0x4c, 0x43, 0x53, 0x79, 0x6e, 0x63,
	0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x2e, 0x41, 0x70, 0x70, 0x54, 0x69, 0x6d, 0x65, 0x41,
	0x6e, 0x73, 0x48, 0x00, 0x52, 0x0a, 0x61, 0x70, 0x70, 0x54, 0x69, 0x6d, 0x65, 0x41, 0x6e, 0x73,
	0x1a, 0x9a, 0x01, 0x0a, 0x0a, 0x41, 0x70, 0x70, 0x54, 0x69, 0x6d, 0x65, 0x52, 0x65, 0x71, 0x12,
	0x44, 0x0a, 0x0a, 0x44, 0x65, 0x76, 0x69, 0x63, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x42,
	0x08, 0xfa, 0x42, 0x05, 0xb2, 0x01, 0x02, 0x08, 0x01, 0x52, 0x0a, 0x44, 0x65, 0x76, 0x69, 0x63,
	0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x24, 0x0a, 0x08, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x52, 0x65,
	0x71, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x42, 0x08, 0xfa, 0x42, 0x05, 0x2a, 0x03, 0x18, 0xff,
	0x01, 0x52, 0x08, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x52, 0x65, 0x71, 0x12, 0x20, 0x0a, 0x0b, 0x41,
	0x6e, 0x73, 0x52, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x0b, 0x41, 0x6e, 0x73, 0x52, 0x65, 0x71, 0x75, 0x69, 0x72, 0x65, 0x64, 0x1a, 0x5a, 0x0a,
	0x0a, 0x41, 0x70, 0x70, 0x54, 0x69, 0x6d, 0x65, 0x41, 0x6e, 0x73, 0x12, 0x26, 0x0a, 0x0e, 0x54,
	0x69, 0x6d, 0x65, 0x43, 0x6f, 0x72, 0x72, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x05, 0x52, 0x0e, 0x54, 0x69, 0x6d, 0x65, 0x43, 0x6f, 0x72, 0x72, 0x65, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x12, 0x24, 0x0a, 0x08, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x41, 0x6e, 0x73, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x0d, 0x42, 0x08, 0xfa, 0x42, 0x05, 0x2a, 0x03, 0x18, 0xff, 0x01, 0x52,
	0x08, 0x54, 0x6f, 0x6b, 0x65, 0x6e, 0x41, 0x6e, 0x73, 0x42, 0x09, 0x0a, 0x07, 0x70, 0x61, 0x79,
	0x6c, 0x6f, 0x61, 0x64, 0x2a, 0xb2, 0x01, 0x0a, 0x18, 0x41, 0x4c, 0x43, 0x53, 0x79, 0x6e, 0x63,
	0x43, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x49, 0x64, 0x65, 0x6e, 0x74, 0x69, 0x66, 0x69, 0x65,
	0x72, 0x12, 0x1b, 0x0a, 0x17, 0x41, 0x4c, 0x43, 0x53, 0x59, 0x4e, 0x43, 0x5f, 0x43, 0x49, 0x44,
	0x5f, 0x50, 0x4b, 0x47, 0x5f, 0x56, 0x45, 0x52, 0x53, 0x49, 0x4f, 0x4e, 0x10, 0x00, 0x12, 0x18,
	0x0a, 0x14, 0x41, 0x4c, 0x43, 0x53, 0x59, 0x4e, 0x43, 0x5f, 0x43, 0x49, 0x44, 0x5f, 0x41, 0x50,
	0x50, 0x5f, 0x54, 0x49, 0x4d, 0x45, 0x10, 0x01, 0x12, 0x28, 0x0a, 0x24, 0x41, 0x4c, 0x43, 0x53,
	0x59, 0x4e, 0x43, 0x5f, 0x43, 0x49, 0x44, 0x5f, 0x41, 0x50, 0x50, 0x5f, 0x44, 0x45, 0x56, 0x5f,
	0x54, 0x49, 0x4d, 0x45, 0x5f, 0x50, 0x45, 0x52, 0x49, 0x4f, 0x44, 0x49, 0x43, 0x49, 0x54, 0x59,
	0x10, 0x02, 0x12, 0x20, 0x0a, 0x1c, 0x41, 0x4c, 0x43, 0x53, 0x59, 0x4e, 0x43, 0x5f, 0x43, 0x49,
	0x44, 0x5f, 0x46, 0x4f, 0x52, 0x43, 0x45, 0x5f, 0x44, 0x45, 0x56, 0x5f, 0x52, 0x45, 0x53, 0x59,
	0x4e, 0x43, 0x10, 0x03, 0x1a, 0x13, 0xea, 0xaa, 0x19, 0x0f, 0x18, 0x01, 0x2a, 0x0b, 0x41, 0x4c,
	0x43, 0x53, 0x59, 0x4e, 0x43, 0x5f, 0x43, 0x49, 0x44, 0x42, 0x31, 0x5a, 0x2f, 0x67, 0x6f, 0x2e,
	0x74, 0x68, 0x65, 0x74, 0x68, 0x69, 0x6e, 0x67, 0x73, 0x2e, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72,
	0x6b, 0x2f, 0x6c, 0x6f, 0x72, 0x61, 0x77, 0x61, 0x6e, 0x2d, 0x73, 0x74, 0x61, 0x63, 0x6b, 0x2f,
	0x76, 0x33, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x74, 0x74, 0x6e, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_rawDescOnce sync.Once
	file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_rawDescData = file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_rawDesc
)

func file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_rawDescGZIP() []byte {
	file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_rawDescOnce.Do(func() {
		file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_rawDescData = protoimpl.X.CompressGZIP(file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_rawDescData)
	})
	return file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_rawDescData
}

var file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_goTypes = []interface{}{
	(ALCSyncCommandIdentifier)(0),     // 0: ttn.lorawan.v3.ALCSyncCommandIdentifier
	(*ALCSyncCommand)(nil),            // 1: ttn.lorawan.v3.ALCSyncCommand
	(*ALCSyncCommand_AppTimeReq)(nil), // 2: ttn.lorawan.v3.ALCSyncCommand.AppTimeReq
	(*ALCSyncCommand_AppTimeAns)(nil), // 3: ttn.lorawan.v3.ALCSyncCommand.AppTimeAns
	(*timestamppb.Timestamp)(nil),     // 4: google.protobuf.Timestamp
}
var file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_depIdxs = []int32{
	0, // 0: ttn.lorawan.v3.ALCSyncCommand.cid:type_name -> ttn.lorawan.v3.ALCSyncCommandIdentifier
	2, // 1: ttn.lorawan.v3.ALCSyncCommand.app_time_req:type_name -> ttn.lorawan.v3.ALCSyncCommand.AppTimeReq
	3, // 2: ttn.lorawan.v3.ALCSyncCommand.app_time_ans:type_name -> ttn.lorawan.v3.ALCSyncCommand.AppTimeAns
	4, // 3: ttn.lorawan.v3.ALCSyncCommand.AppTimeReq.DeviceTime:type_name -> google.protobuf.Timestamp
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_init() }
func file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_init() {
	if File_lorawan_stack_api_applicationserver_integrations_alcsync_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ALCSyncCommand); i {
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
		file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ALCSyncCommand_AppTimeReq); i {
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
		file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ALCSyncCommand_AppTimeAns); i {
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
	file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*ALCSyncCommand_AppTimeReq_)(nil),
		(*ALCSyncCommand_AppTimeAns_)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_goTypes,
		DependencyIndexes: file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_depIdxs,
		EnumInfos:         file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_enumTypes,
		MessageInfos:      file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_msgTypes,
	}.Build()
	File_lorawan_stack_api_applicationserver_integrations_alcsync_proto = out.File
	file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_rawDesc = nil
	file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_goTypes = nil
	file_lorawan_stack_api_applicationserver_integrations_alcsync_proto_depIdxs = nil
}
