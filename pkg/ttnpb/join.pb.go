// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: lorawan-stack/api/join.proto

package ttnpb

import (
	bytes "bytes"
	fmt "fmt"
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/gogo/protobuf/types"
	github_com_gogo_protobuf_types "github.com/gogo/protobuf/types"
	golang_proto "github.com/golang/protobuf/proto"
	go_thethings_network_lorawan_stack_v3_pkg_types "go.thethings.network/lorawan-stack/v3/pkg/types"
	io "io"
	math "math"
	math_bits "math/bits"
	reflect "reflect"
	strings "strings"
	time "time"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = golang_proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

type JoinRequest struct {
	RawPayload         []byte                                                  `protobuf:"bytes,1,opt,name=raw_payload,json=rawPayload,proto3" json:"raw_payload,omitempty"`
	Payload            *Message                                                `protobuf:"bytes,2,opt,name=payload,proto3" json:"payload,omitempty"`
	DevAddr            go_thethings_network_lorawan_stack_v3_pkg_types.DevAddr `protobuf:"bytes,3,opt,name=dev_addr,json=devAddr,proto3,customtype=go.thethings.network/lorawan-stack/v3/pkg/types.DevAddr" json:"dev_addr"`
	SelectedMACVersion MACVersion                                              `protobuf:"varint,4,opt,name=selected_mac_version,json=selectedMacVersion,proto3,enum=ttn.lorawan.v3.MACVersion" json:"selected_mac_version,omitempty"`
	NetID              go_thethings_network_lorawan_stack_v3_pkg_types.NetID   `protobuf:"bytes,5,opt,name=net_id,json=netId,proto3,customtype=go.thethings.network/lorawan-stack/v3/pkg/types.NetID" json:"net_id"`
	DownlinkSettings   DLSettings                                              `protobuf:"bytes,6,opt,name=downlink_settings,json=downlinkSettings,proto3" json:"downlink_settings"`
	RxDelay            RxDelay                                                 `protobuf:"varint,7,opt,name=rx_delay,json=rxDelay,proto3,enum=ttn.lorawan.v3.RxDelay" json:"rx_delay,omitempty"`
	// Optional CFList.
	CFList         *CFList  `protobuf:"bytes,8,opt,name=cf_list,json=cfList,proto3" json:"cf_list,omitempty"`
	CorrelationIDs []string `protobuf:"bytes,10,rep,name=correlation_ids,json=correlationIds,proto3" json:"correlation_ids,omitempty"`
	// Consumed airtime for the transmission of the join request. Calculated by Network Server using the RawPayload size and the transmission settings.
	ConsumedAirtime      *time.Duration `protobuf:"bytes,11,opt,name=consumed_airtime,json=consumedAirtime,proto3,stdduration" json:"consumed_airtime,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *JoinRequest) Reset()      { *m = JoinRequest{} }
func (*JoinRequest) ProtoMessage() {}
func (*JoinRequest) Descriptor() ([]byte, []int) {
	return fileDescriptor_dd69b88666e72e14, []int{0}
}
func (m *JoinRequest) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *JoinRequest) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_JoinRequest.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *JoinRequest) XXX_Merge(src proto.Message) {
	xxx_messageInfo_JoinRequest.Merge(m, src)
}
func (m *JoinRequest) XXX_Size() int {
	return m.Size()
}
func (m *JoinRequest) XXX_DiscardUnknown() {
	xxx_messageInfo_JoinRequest.DiscardUnknown(m)
}

var xxx_messageInfo_JoinRequest proto.InternalMessageInfo

func (m *JoinRequest) GetRawPayload() []byte {
	if m != nil {
		return m.RawPayload
	}
	return nil
}

func (m *JoinRequest) GetPayload() *Message {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (m *JoinRequest) GetSelectedMACVersion() MACVersion {
	if m != nil {
		return m.SelectedMACVersion
	}
	return MAC_UNKNOWN
}

func (m *JoinRequest) GetDownlinkSettings() DLSettings {
	if m != nil {
		return m.DownlinkSettings
	}
	return DLSettings{}
}

func (m *JoinRequest) GetRxDelay() RxDelay {
	if m != nil {
		return m.RxDelay
	}
	return RX_DELAY_0
}

func (m *JoinRequest) GetCFList() *CFList {
	if m != nil {
		return m.CFList
	}
	return nil
}

func (m *JoinRequest) GetCorrelationIDs() []string {
	if m != nil {
		return m.CorrelationIDs
	}
	return nil
}

func (m *JoinRequest) GetConsumedAirtime() *time.Duration {
	if m != nil {
		return m.ConsumedAirtime
	}
	return nil
}

type JoinResponse struct {
	RawPayload           []byte `protobuf:"bytes,1,opt,name=raw_payload,json=rawPayload,proto3" json:"raw_payload,omitempty"`
	SessionKeys          `protobuf:"bytes,2,opt,name=session_keys,json=sessionKeys,proto3,embedded=session_keys" json:"session_keys"`
	Lifetime             time.Duration `protobuf:"bytes,3,opt,name=lifetime,proto3,stdduration" json:"lifetime"`
	CorrelationIDs       []string      `protobuf:"bytes,4,rep,name=correlation_ids,json=correlationIds,proto3" json:"correlation_ids,omitempty"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *JoinResponse) Reset()      { *m = JoinResponse{} }
func (*JoinResponse) ProtoMessage() {}
func (*JoinResponse) Descriptor() ([]byte, []int) {
	return fileDescriptor_dd69b88666e72e14, []int{1}
}
func (m *JoinResponse) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *JoinResponse) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_JoinResponse.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *JoinResponse) XXX_Merge(src proto.Message) {
	xxx_messageInfo_JoinResponse.Merge(m, src)
}
func (m *JoinResponse) XXX_Size() int {
	return m.Size()
}
func (m *JoinResponse) XXX_DiscardUnknown() {
	xxx_messageInfo_JoinResponse.DiscardUnknown(m)
}

var xxx_messageInfo_JoinResponse proto.InternalMessageInfo

func (m *JoinResponse) GetRawPayload() []byte {
	if m != nil {
		return m.RawPayload
	}
	return nil
}

func (m *JoinResponse) GetLifetime() time.Duration {
	if m != nil {
		return m.Lifetime
	}
	return 0
}

func (m *JoinResponse) GetCorrelationIDs() []string {
	if m != nil {
		return m.CorrelationIDs
	}
	return nil
}

func init() {
	proto.RegisterType((*JoinRequest)(nil), "ttn.lorawan.v3.JoinRequest")
	golang_proto.RegisterType((*JoinRequest)(nil), "ttn.lorawan.v3.JoinRequest")
	proto.RegisterType((*JoinResponse)(nil), "ttn.lorawan.v3.JoinResponse")
	golang_proto.RegisterType((*JoinResponse)(nil), "ttn.lorawan.v3.JoinResponse")
}

func init() { proto.RegisterFile("lorawan-stack/api/join.proto", fileDescriptor_dd69b88666e72e14) }
func init() {
	golang_proto.RegisterFile("lorawan-stack/api/join.proto", fileDescriptor_dd69b88666e72e14)
}

var fileDescriptor_dd69b88666e72e14 = []byte{
	// 771 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x53, 0x31, 0x6f, 0xdb, 0x46,
	0x18, 0xd5, 0xc9, 0x12, 0x25, 0x9f, 0x0c, 0x47, 0x21, 0x8a, 0x84, 0x75, 0x0b, 0xd2, 0xf5, 0x24,
	0x14, 0x30, 0x89, 0xc6, 0x28, 0x0a, 0xb4, 0x05, 0x02, 0xd3, 0x42, 0x0a, 0xa7, 0x49, 0x10, 0xd0,
	0x68, 0x87, 0x00, 0x05, 0x71, 0xe6, 0x9d, 0xa8, 0xab, 0xa8, 0x3b, 0xf6, 0xee, 0x24, 0x59, 0x99,
	0x8a, 0x8c, 0x9d, 0x8a, 0x0e, 0x45, 0xc7, 0x8c, 0xfd, 0x09, 0x1d, 0x3b, 0x7a, 0xcc, 0x58, 0x74,
	0x60, 0x1b, 0x7a, 0xe9, 0xd8, 0xb1, 0xd0, 0x54, 0xe8, 0x48, 0x56, 0x76, 0xe4, 0xa1, 0xc9, 0xc4,
	0x8f, 0xf7, 0xbd, 0xf7, 0xf0, 0xbe, 0xef, 0xee, 0xc1, 0x77, 0x13, 0x2e, 0xd0, 0x0c, 0xb1, 0x7d,
	0xa9, 0x50, 0x34, 0xf2, 0x50, 0x4a, 0xbd, 0xaf, 0x39, 0x65, 0x6e, 0x2a, 0xb8, 0xe2, 0xe6, 0xb6,
	0x52, 0xcc, 0x2d, 0x11, 0xee, 0xf4, 0x60, 0xe7, 0x30, 0xa6, 0x6a, 0x38, 0x39, 0x75, 0x23, 0x3e,
	0xf6, 0x08, 0x9b, 0xf2, 0x79, 0x2a, 0xf8, 0xd9, 0xdc, 0xd3, 0xe0, 0x68, 0x3f, 0x26, 0x6c, 0x7f,
	0x8a, 0x12, 0x8a, 0x91, 0x22, 0xde, 0x5a, 0x51, 0x48, 0xee, 0xec, 0x5f, 0x92, 0x88, 0x79, 0xcc,
	0x0b, 0xf2, 0xe9, 0x64, 0xa0, 0xff, 0xf4, 0x8f, 0xae, 0x4a, 0xb8, 0x1d, 0x73, 0x1e, 0x27, 0x64,
	0x85, 0xc2, 0x13, 0x81, 0x14, 0xe5, 0xa5, 0xc3, 0x9d, 0x6b, 0xfc, 0x8f, 0xc8, 0x5c, 0x96, 0x5d,
	0x67, 0xbd, 0x5b, 0x4d, 0xa3, 0x01, 0x7b, 0xcf, 0x0c, 0xd8, 0xb9, 0xcf, 0x29, 0x0b, 0xc8, 0x37,
	0x13, 0x22, 0x95, 0xd9, 0x83, 0x1d, 0x81, 0x66, 0x61, 0x8a, 0xe6, 0x09, 0x47, 0xd8, 0x02, 0xbb,
	0xa0, 0xb7, 0xe5, 0xb7, 0x16, 0x7e, 0xe3, 0x69, 0x7d, 0x78, 0x3b, 0x80, 0x02, 0xcd, 0x1e, 0x17,
	0x2d, 0xf3, 0x03, 0xd8, 0xaa, 0x50, 0xf5, 0x5d, 0xd0, 0xeb, 0xdc, 0xb9, 0xed, 0x5e, 0x5d, 0x96,
	0xfb, 0x90, 0x48, 0x89, 0x62, 0x12, 0x54, 0x38, 0xf3, 0x09, 0x6c, 0x63, 0x32, 0x0d, 0x11, 0xc6,
	0xc2, 0xda, 0xd0, 0xca, 0x77, 0xcf, 0x33, 0xa7, 0xf6, 0x7b, 0xe6, 0x7c, 0x14, 0x73, 0x57, 0x0d,
	0x89, 0x1a, 0x52, 0x16, 0x4b, 0x97, 0x11, 0x35, 0xe3, 0x62, 0xe4, 0x5d, 0x35, 0x3f, 0x3d, 0xf0,
	0xd2, 0x51, 0xec, 0xa9, 0x79, 0x4a, 0xa4, 0xdb, 0x27, 0xd3, 0x43, 0x8c, 0x45, 0xd0, 0xc2, 0x45,
	0x61, 0x62, 0xf8, 0x96, 0x24, 0x09, 0x89, 0x14, 0xc1, 0xe1, 0x18, 0x45, 0xe1, 0x94, 0x08, 0x49,
	0x39, 0xb3, 0x1a, 0xbb, 0xa0, 0xb7, 0x7d, 0x67, 0x67, 0xcd, 0xdb, 0xe1, 0xd1, 0x97, 0x05, 0xc2,
	0xbf, 0x95, 0x67, 0x8e, 0x79, 0x52, 0x72, 0x57, 0xe7, 0x81, 0x59, 0xe9, 0x3d, 0x44, 0x51, 0x79,
	0x66, 0x7e, 0x05, 0x0d, 0x46, 0x54, 0x48, 0xb1, 0xd5, 0xd4, 0xfe, 0xef, 0x95, 0xfe, 0x3f, 0x7c,
	0x5d, 0xff, 0x8f, 0x88, 0x3a, 0xee, 0xe7, 0x99, 0xd3, 0xd4, 0x45, 0xd0, 0x64, 0x44, 0x1d, 0x63,
	0xf3, 0x0b, 0x78, 0x13, 0xf3, 0x19, 0x4b, 0x28, 0x1b, 0x85, 0x92, 0x28, 0xb5, 0x54, 0xb3, 0x0c,
	0xbd, 0xdd, 0xb5, 0x09, 0xfa, 0x0f, 0x4e, 0x4a, 0x84, 0xbf, 0xb5, 0xf0, 0x9b, 0xdf, 0x81, 0x7a,
	0x17, 0x2c, 0xdd, 0x04, 0xdd, 0x4a, 0xa2, 0xea, 0x9b, 0x9f, 0xc2, 0xb6, 0x38, 0x0b, 0x31, 0x49,
	0xd0, 0xdc, 0x6a, 0xe9, 0x7d, 0xac, 0xdd, 0x55, 0x70, 0xd6, 0x5f, 0xb6, 0xfd, 0xf6, 0xc2, 0x6f,
	0x3e, 0x5b, 0x4a, 0x05, 0x2d, 0x51, 0x1c, 0x99, 0x9f, 0xc0, 0x56, 0x34, 0x08, 0x13, 0x2a, 0x95,
	0xd5, 0xd6, 0x56, 0x6e, 0xbd, 0x4a, 0x3e, 0xba, 0xf7, 0x80, 0x4a, 0xe5, 0xc3, 0x3c, 0x73, 0x8c,
	0xa2, 0x0e, 0x8c, 0x68, 0xb0, 0xfc, 0x9a, 0x9f, 0xc1, 0x1b, 0x11, 0x17, 0x82, 0x24, 0xfa, 0xcd,
	0x86, 0x14, 0x4b, 0x0b, 0xee, 0x6e, 0xf4, 0x36, 0x7d, 0x7b, 0xe1, 0x6f, 0xfe, 0x00, 0x8c, 0xbd,
	0x86, 0xa8, 0x5b, 0x38, 0xcf, 0x9c, 0xed, 0xa3, 0x15, 0xec, 0xb8, 0x2f, 0x83, 0xed, 0x4b, 0xb4,
	0x63, 0x2c, 0xcd, 0x47, 0xb0, 0x1b, 0x71, 0x26, 0x27, 0x63, 0x82, 0x43, 0x44, 0x85, 0xa2, 0x63,
	0x62, 0x75, 0xb4, 0x9d, 0xb7, 0xdd, 0x22, 0x22, 0x6e, 0x15, 0x11, 0xb7, 0x5f, 0x46, 0xc4, 0x6f,
	0x9f, 0x67, 0x0e, 0xf8, 0xe9, 0x0f, 0x07, 0x04, 0x37, 0x2a, 0xf2, 0x61, 0xc1, 0xfd, 0xb8, 0xf1,
	0xcb, 0x73, 0xa7, 0x76, 0xbf, 0xd1, 0xde, 0xec, 0xc2, 0xbd, 0x1f, 0xeb, 0x70, 0xab, 0x08, 0x81,
	0x4c, 0x39, 0x93, 0xc4, 0x7c, 0xff, 0xba, 0x14, 0x6c, 0x2e, 0x7c, 0xe3, 0x69, 0xa3, 0x7b, 0xd3,
	0x7a, 0xef, 0x4a, 0x0e, 0x1e, 0xc3, 0x2d, 0x49, 0xe4, 0xf2, 0x75, 0x84, 0xcb, 0xe0, 0x95, 0x61,
	0x78, 0xe7, 0xd5, 0x1d, 0x9d, 0x14, 0x98, 0xcf, 0xc9, 0x5c, 0xfa, 0xdd, 0xcb, 0xf7, 0xf5, 0x22,
	0x73, 0x40, 0xd0, 0x91, 0xab, 0xb6, 0x79, 0x17, 0xb6, 0x13, 0x3a, 0x20, 0x7a, 0xc4, 0x8d, 0xff,
	0x33, 0x62, 0x4d, 0x8f, 0xf8, 0x1f, 0xe9, 0xba, 0xa5, 0x37, 0xde, 0x64, 0xe9, 0x3e, 0xfa, 0xed,
	0xa5, 0x5d, 0xfb, 0xe7, 0xa5, 0x0d, 0xbe, 0xcd, 0x6d, 0xf0, 0x73, 0x6e, 0x83, 0xf3, 0xdc, 0x06,
	0x2f, 0x72, 0x1b, 0xfc, 0x99, 0xdb, 0xe0, 0xaf, 0xdc, 0xae, 0xfd, 0x9d, 0xdb, 0xe0, 0xfb, 0x0b,
	0xbb, 0xf6, 0xfc, 0xc2, 0xae, 0xfd, 0x7a, 0x61, 0x83, 0x27, 0xde, 0x6b, 0x04, 0x41, 0xb1, 0xf4,
	0xf4, 0xd4, 0xd0, 0x23, 0x1d, 0xfc, 0x1b, 0x00, 0x00, 0xff, 0xff, 0x85, 0x0f, 0x71, 0xc3, 0x88,
	0x05, 0x00, 0x00,
}

func (this *JoinRequest) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*JoinRequest)
	if !ok {
		that2, ok := that.(JoinRequest)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if !bytes.Equal(this.RawPayload, that1.RawPayload) {
		return false
	}
	if !this.Payload.Equal(that1.Payload) {
		return false
	}
	if !this.DevAddr.Equal(that1.DevAddr) {
		return false
	}
	if this.SelectedMACVersion != that1.SelectedMACVersion {
		return false
	}
	if !this.NetID.Equal(that1.NetID) {
		return false
	}
	if !this.DownlinkSettings.Equal(&that1.DownlinkSettings) {
		return false
	}
	if this.RxDelay != that1.RxDelay {
		return false
	}
	if !this.CFList.Equal(that1.CFList) {
		return false
	}
	if len(this.CorrelationIDs) != len(that1.CorrelationIDs) {
		return false
	}
	for i := range this.CorrelationIDs {
		if this.CorrelationIDs[i] != that1.CorrelationIDs[i] {
			return false
		}
	}
	if this.ConsumedAirtime != nil && that1.ConsumedAirtime != nil {
		if *this.ConsumedAirtime != *that1.ConsumedAirtime {
			return false
		}
	} else if this.ConsumedAirtime != nil {
		return false
	} else if that1.ConsumedAirtime != nil {
		return false
	}
	return true
}
func (this *JoinResponse) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*JoinResponse)
	if !ok {
		that2, ok := that.(JoinResponse)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if !bytes.Equal(this.RawPayload, that1.RawPayload) {
		return false
	}
	if !this.SessionKeys.Equal(&that1.SessionKeys) {
		return false
	}
	if this.Lifetime != that1.Lifetime {
		return false
	}
	if len(this.CorrelationIDs) != len(that1.CorrelationIDs) {
		return false
	}
	for i := range this.CorrelationIDs {
		if this.CorrelationIDs[i] != that1.CorrelationIDs[i] {
			return false
		}
	}
	return true
}
func (m *JoinRequest) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *JoinRequest) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *JoinRequest) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.ConsumedAirtime != nil {
		n1, err1 := github_com_gogo_protobuf_types.StdDurationMarshalTo(*m.ConsumedAirtime, dAtA[i-github_com_gogo_protobuf_types.SizeOfStdDuration(*m.ConsumedAirtime):])
		if err1 != nil {
			return 0, err1
		}
		i -= n1
		i = encodeVarintJoin(dAtA, i, uint64(n1))
		i--
		dAtA[i] = 0x5a
	}
	if len(m.CorrelationIDs) > 0 {
		for iNdEx := len(m.CorrelationIDs) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.CorrelationIDs[iNdEx])
			copy(dAtA[i:], m.CorrelationIDs[iNdEx])
			i = encodeVarintJoin(dAtA, i, uint64(len(m.CorrelationIDs[iNdEx])))
			i--
			dAtA[i] = 0x52
		}
	}
	if m.CFList != nil {
		{
			size, err := m.CFList.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintJoin(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x42
	}
	if m.RxDelay != 0 {
		i = encodeVarintJoin(dAtA, i, uint64(m.RxDelay))
		i--
		dAtA[i] = 0x38
	}
	{
		size, err := m.DownlinkSettings.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintJoin(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x32
	{
		size := m.NetID.Size()
		i -= size
		if _, err := m.NetID.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintJoin(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x2a
	if m.SelectedMACVersion != 0 {
		i = encodeVarintJoin(dAtA, i, uint64(m.SelectedMACVersion))
		i--
		dAtA[i] = 0x20
	}
	{
		size := m.DevAddr.Size()
		i -= size
		if _, err := m.DevAddr.MarshalTo(dAtA[i:]); err != nil {
			return 0, err
		}
		i = encodeVarintJoin(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x1a
	if m.Payload != nil {
		{
			size, err := m.Payload.MarshalToSizedBuffer(dAtA[:i])
			if err != nil {
				return 0, err
			}
			i -= size
			i = encodeVarintJoin(dAtA, i, uint64(size))
		}
		i--
		dAtA[i] = 0x12
	}
	if len(m.RawPayload) > 0 {
		i -= len(m.RawPayload)
		copy(dAtA[i:], m.RawPayload)
		i = encodeVarintJoin(dAtA, i, uint64(len(m.RawPayload)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func (m *JoinResponse) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *JoinResponse) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *JoinResponse) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.CorrelationIDs) > 0 {
		for iNdEx := len(m.CorrelationIDs) - 1; iNdEx >= 0; iNdEx-- {
			i -= len(m.CorrelationIDs[iNdEx])
			copy(dAtA[i:], m.CorrelationIDs[iNdEx])
			i = encodeVarintJoin(dAtA, i, uint64(len(m.CorrelationIDs[iNdEx])))
			i--
			dAtA[i] = 0x22
		}
	}
	n5, err5 := github_com_gogo_protobuf_types.StdDurationMarshalTo(m.Lifetime, dAtA[i-github_com_gogo_protobuf_types.SizeOfStdDuration(m.Lifetime):])
	if err5 != nil {
		return 0, err5
	}
	i -= n5
	i = encodeVarintJoin(dAtA, i, uint64(n5))
	i--
	dAtA[i] = 0x1a
	{
		size, err := m.SessionKeys.MarshalToSizedBuffer(dAtA[:i])
		if err != nil {
			return 0, err
		}
		i -= size
		i = encodeVarintJoin(dAtA, i, uint64(size))
	}
	i--
	dAtA[i] = 0x12
	if len(m.RawPayload) > 0 {
		i -= len(m.RawPayload)
		copy(dAtA[i:], m.RawPayload)
		i = encodeVarintJoin(dAtA, i, uint64(len(m.RawPayload)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintJoin(dAtA []byte, offset int, v uint64) int {
	offset -= sovJoin(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func NewPopulatedJoinResponse(r randyJoin, easy bool) *JoinResponse {
	this := &JoinResponse{}
	v1 := r.Intn(100)
	this.RawPayload = make([]byte, v1)
	for i := 0; i < v1; i++ {
		this.RawPayload[i] = byte(r.Intn(256))
	}
	v2 := NewPopulatedSessionKeys(r, easy)
	this.SessionKeys = *v2
	v3 := github_com_gogo_protobuf_types.NewPopulatedStdDuration(r, easy)
	this.Lifetime = *v3
	v4 := r.Intn(10)
	this.CorrelationIDs = make([]string, v4)
	for i := 0; i < v4; i++ {
		this.CorrelationIDs[i] = string(randStringJoin(r))
	}
	if !easy && r.Intn(10) != 0 {
	}
	return this
}

type randyJoin interface {
	Float32() float32
	Float64() float64
	Int63() int64
	Int31() int32
	Uint32() uint32
	Intn(n int) int
}

func randUTF8RuneJoin(r randyJoin) rune {
	ru := r.Intn(62)
	if ru < 10 {
		return rune(ru + 48)
	} else if ru < 36 {
		return rune(ru + 55)
	}
	return rune(ru + 61)
}
func randStringJoin(r randyJoin) string {
	v5 := r.Intn(100)
	tmps := make([]rune, v5)
	for i := 0; i < v5; i++ {
		tmps[i] = randUTF8RuneJoin(r)
	}
	return string(tmps)
}
func randUnrecognizedJoin(r randyJoin, maxFieldNumber int) (dAtA []byte) {
	l := r.Intn(5)
	for i := 0; i < l; i++ {
		wire := r.Intn(4)
		if wire == 3 {
			wire = 5
		}
		fieldNumber := maxFieldNumber + r.Intn(100)
		dAtA = randFieldJoin(dAtA, r, fieldNumber, wire)
	}
	return dAtA
}
func randFieldJoin(dAtA []byte, r randyJoin, fieldNumber int, wire int) []byte {
	key := uint32(fieldNumber)<<3 | uint32(wire)
	switch wire {
	case 0:
		dAtA = encodeVarintPopulateJoin(dAtA, uint64(key))
		v6 := r.Int63()
		if r.Intn(2) == 0 {
			v6 *= -1
		}
		dAtA = encodeVarintPopulateJoin(dAtA, uint64(v6))
	case 1:
		dAtA = encodeVarintPopulateJoin(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	case 2:
		dAtA = encodeVarintPopulateJoin(dAtA, uint64(key))
		ll := r.Intn(100)
		dAtA = encodeVarintPopulateJoin(dAtA, uint64(ll))
		for j := 0; j < ll; j++ {
			dAtA = append(dAtA, byte(r.Intn(256)))
		}
	default:
		dAtA = encodeVarintPopulateJoin(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	}
	return dAtA
}
func encodeVarintPopulateJoin(dAtA []byte, v uint64) []byte {
	for v >= 1<<7 {
		dAtA = append(dAtA, uint8(uint64(v)&0x7f|0x80))
		v >>= 7
	}
	dAtA = append(dAtA, uint8(v))
	return dAtA
}
func (m *JoinRequest) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.RawPayload)
	if l > 0 {
		n += 1 + l + sovJoin(uint64(l))
	}
	if m.Payload != nil {
		l = m.Payload.Size()
		n += 1 + l + sovJoin(uint64(l))
	}
	l = m.DevAddr.Size()
	n += 1 + l + sovJoin(uint64(l))
	if m.SelectedMACVersion != 0 {
		n += 1 + sovJoin(uint64(m.SelectedMACVersion))
	}
	l = m.NetID.Size()
	n += 1 + l + sovJoin(uint64(l))
	l = m.DownlinkSettings.Size()
	n += 1 + l + sovJoin(uint64(l))
	if m.RxDelay != 0 {
		n += 1 + sovJoin(uint64(m.RxDelay))
	}
	if m.CFList != nil {
		l = m.CFList.Size()
		n += 1 + l + sovJoin(uint64(l))
	}
	if len(m.CorrelationIDs) > 0 {
		for _, s := range m.CorrelationIDs {
			l = len(s)
			n += 1 + l + sovJoin(uint64(l))
		}
	}
	if m.ConsumedAirtime != nil {
		l = github_com_gogo_protobuf_types.SizeOfStdDuration(*m.ConsumedAirtime)
		n += 1 + l + sovJoin(uint64(l))
	}
	return n
}

func (m *JoinResponse) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.RawPayload)
	if l > 0 {
		n += 1 + l + sovJoin(uint64(l))
	}
	l = m.SessionKeys.Size()
	n += 1 + l + sovJoin(uint64(l))
	l = github_com_gogo_protobuf_types.SizeOfStdDuration(m.Lifetime)
	n += 1 + l + sovJoin(uint64(l))
	if len(m.CorrelationIDs) > 0 {
		for _, s := range m.CorrelationIDs {
			l = len(s)
			n += 1 + l + sovJoin(uint64(l))
		}
	}
	return n
}

func sovJoin(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozJoin(x uint64) (n int) {
	return sovJoin(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *JoinRequest) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&JoinRequest{`,
		`RawPayload:` + fmt.Sprintf("%v", this.RawPayload) + `,`,
		`Payload:` + strings.Replace(fmt.Sprintf("%v", this.Payload), "Message", "Message", 1) + `,`,
		`DevAddr:` + fmt.Sprintf("%v", this.DevAddr) + `,`,
		`SelectedMACVersion:` + fmt.Sprintf("%v", this.SelectedMACVersion) + `,`,
		`NetID:` + fmt.Sprintf("%v", this.NetID) + `,`,
		`DownlinkSettings:` + strings.Replace(strings.Replace(fmt.Sprintf("%v", this.DownlinkSettings), "DLSettings", "DLSettings", 1), `&`, ``, 1) + `,`,
		`RxDelay:` + fmt.Sprintf("%v", this.RxDelay) + `,`,
		`CFList:` + strings.Replace(fmt.Sprintf("%v", this.CFList), "CFList", "CFList", 1) + `,`,
		`CorrelationIDs:` + fmt.Sprintf("%v", this.CorrelationIDs) + `,`,
		`ConsumedAirtime:` + strings.Replace(fmt.Sprintf("%v", this.ConsumedAirtime), "Duration", "types.Duration", 1) + `,`,
		`}`,
	}, "")
	return s
}
func (this *JoinResponse) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&JoinResponse{`,
		`RawPayload:` + fmt.Sprintf("%v", this.RawPayload) + `,`,
		`SessionKeys:` + strings.Replace(strings.Replace(fmt.Sprintf("%v", this.SessionKeys), "SessionKeys", "SessionKeys", 1), `&`, ``, 1) + `,`,
		`Lifetime:` + strings.Replace(strings.Replace(fmt.Sprintf("%v", this.Lifetime), "Duration", "types.Duration", 1), `&`, ``, 1) + `,`,
		`CorrelationIDs:` + fmt.Sprintf("%v", this.CorrelationIDs) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringJoin(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *JoinRequest) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowJoin
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: JoinRequest: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: JoinRequest: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RawPayload", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowJoin
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthJoin
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthJoin
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.RawPayload = append(m.RawPayload[:0], dAtA[iNdEx:postIndex]...)
			if m.RawPayload == nil {
				m.RawPayload = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Payload", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowJoin
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthJoin
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthJoin
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.Payload == nil {
				m.Payload = &Message{}
			}
			if err := m.Payload.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DevAddr", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowJoin
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthJoin
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthJoin
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.DevAddr.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field SelectedMACVersion", wireType)
			}
			m.SelectedMACVersion = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowJoin
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.SelectedMACVersion |= MACVersion(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field NetID", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowJoin
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthJoin
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthJoin
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.NetID.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 6:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field DownlinkSettings", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowJoin
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthJoin
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthJoin
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.DownlinkSettings.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 7:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field RxDelay", wireType)
			}
			m.RxDelay = 0
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowJoin
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				m.RxDelay |= RxDelay(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
		case 8:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CFList", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowJoin
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthJoin
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthJoin
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.CFList == nil {
				m.CFList = &CFList{}
			}
			if err := m.CFList.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 10:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CorrelationIDs", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowJoin
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthJoin
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthJoin
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.CorrelationIDs = append(m.CorrelationIDs, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		case 11:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field ConsumedAirtime", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowJoin
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthJoin
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthJoin
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.ConsumedAirtime == nil {
				m.ConsumedAirtime = new(time.Duration)
			}
			if err := github_com_gogo_protobuf_types.StdDurationUnmarshal(m.ConsumedAirtime, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipJoin(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthJoin
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthJoin
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func (m *JoinResponse) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowJoin
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: JoinResponse: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: JoinResponse: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field RawPayload", wireType)
			}
			var byteLen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowJoin
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				byteLen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if byteLen < 0 {
				return ErrInvalidLengthJoin
			}
			postIndex := iNdEx + byteLen
			if postIndex < 0 {
				return ErrInvalidLengthJoin
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.RawPayload = append(m.RawPayload[:0], dAtA[iNdEx:postIndex]...)
			if m.RawPayload == nil {
				m.RawPayload = []byte{}
			}
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SessionKeys", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowJoin
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthJoin
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthJoin
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := m.SessionKeys.Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Lifetime", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowJoin
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthJoin
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthJoin
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if err := github_com_gogo_protobuf_types.StdDurationUnmarshal(&m.Lifetime, dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		case 4:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field CorrelationIDs", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowJoin
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthJoin
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthJoin
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.CorrelationIDs = append(m.CorrelationIDs, string(dAtA[iNdEx:postIndex]))
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipJoin(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthJoin
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthJoin
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipJoin(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowJoin
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowJoin
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowJoin
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthJoin
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupJoin
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthJoin
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthJoin        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowJoin          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupJoin = fmt.Errorf("proto: unexpected end of group")
)
