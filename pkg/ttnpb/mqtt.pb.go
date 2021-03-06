// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: lorawan-stack/api/mqtt.proto

package ttnpb

import (
	fmt "fmt"
	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	golang_proto "github.com/golang/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
	reflect "reflect"
	strings "strings"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = golang_proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// The connection information of an MQTT frontend.
type MQTTConnectionInfo struct {
	// The public listen address of the frontend.
	PublicAddress string `protobuf:"bytes,1,opt,name=public_address,json=publicAddress,proto3" json:"public_address,omitempty"`
	// The public listen address of the TLS frontend.
	PublicTLSAddress string `protobuf:"bytes,2,opt,name=public_tls_address,json=publicTlsAddress,proto3" json:"public_tls_address,omitempty"`
	// The username to be used for authentication.
	Username             string   `protobuf:"bytes,3,opt,name=username,proto3" json:"username,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MQTTConnectionInfo) Reset()      { *m = MQTTConnectionInfo{} }
func (*MQTTConnectionInfo) ProtoMessage() {}
func (*MQTTConnectionInfo) Descriptor() ([]byte, []int) {
	return fileDescriptor_dbbf9b6b10797b61, []int{0}
}
func (m *MQTTConnectionInfo) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *MQTTConnectionInfo) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_MQTTConnectionInfo.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *MQTTConnectionInfo) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MQTTConnectionInfo.Merge(m, src)
}
func (m *MQTTConnectionInfo) XXX_Size() int {
	return m.Size()
}
func (m *MQTTConnectionInfo) XXX_DiscardUnknown() {
	xxx_messageInfo_MQTTConnectionInfo.DiscardUnknown(m)
}

var xxx_messageInfo_MQTTConnectionInfo proto.InternalMessageInfo

func (m *MQTTConnectionInfo) GetPublicAddress() string {
	if m != nil {
		return m.PublicAddress
	}
	return ""
}

func (m *MQTTConnectionInfo) GetPublicTLSAddress() string {
	if m != nil {
		return m.PublicTLSAddress
	}
	return ""
}

func (m *MQTTConnectionInfo) GetUsername() string {
	if m != nil {
		return m.Username
	}
	return ""
}

func init() {
	proto.RegisterType((*MQTTConnectionInfo)(nil), "ttn.lorawan.v3.MQTTConnectionInfo")
	golang_proto.RegisterType((*MQTTConnectionInfo)(nil), "ttn.lorawan.v3.MQTTConnectionInfo")
}

func init() { proto.RegisterFile("lorawan-stack/api/mqtt.proto", fileDescriptor_dbbf9b6b10797b61) }
func init() {
	golang_proto.RegisterFile("lorawan-stack/api/mqtt.proto", fileDescriptor_dbbf9b6b10797b61)
}

var fileDescriptor_dbbf9b6b10797b61 = []byte{
	// 425 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0xc9, 0xc9, 0x2f, 0x4a,
	0x2c, 0x4f, 0xcc, 0xd3, 0x2d, 0x2e, 0x49, 0x4c, 0xce, 0xd6, 0x4f, 0x2c, 0xc8, 0xd4, 0xcf, 0x2d,
	0x2c, 0x29, 0xd1, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x2b, 0x29, 0xc9, 0xd3, 0x83, 0xaa,
	0xd0, 0x2b, 0x33, 0x96, 0x72, 0x4c, 0xcf, 0x2c, 0xc9, 0x28, 0x4d, 0xd2, 0x4b, 0xce, 0xcf, 0xd5,
	0x4f, 0xcd, 0x2b, 0xcb, 0xaf, 0x2c, 0x28, 0xca, 0xaf, 0xa8, 0xd4, 0x07, 0x2b, 0x4e, 0xd6, 0x4d,
	0x4f, 0xcd, 0xd3, 0x2d, 0x4b, 0xcc, 0xc9, 0x4c, 0x49, 0x2c, 0x49, 0xd5, 0xc7, 0x60, 0x40, 0x8c,
	0x94, 0xd2, 0x45, 0x32, 0x22, 0x3d, 0x3f, 0x3d, 0x1f, 0xa2, 0x39, 0xa9, 0x34, 0x0d, 0xcc, 0x03,
	0x73, 0xc0, 0x2c, 0x88, 0x72, 0xa5, 0xbd, 0xcc, 0x5c, 0x42, 0xbe, 0x81, 0x21, 0x21, 0xce, 0xf9,
	0x79, 0x79, 0xa9, 0xc9, 0x25, 0x99, 0xf9, 0x79, 0x9e, 0x79, 0x69, 0xf9, 0x42, 0xdb, 0x18, 0xb9,
	0xf8, 0x0a, 0x4a, 0x93, 0x72, 0x32, 0x93, 0xe3, 0x13, 0x53, 0x52, 0x8a, 0x52, 0x8b, 0x8b, 0x25,
	0x18, 0x15, 0x18, 0x35, 0x38, 0x9d, 0xfa, 0x18, 0x7f, 0x39, 0x75, 0x31, 0x16, 0xb5, 0x33, 0x1a,
	0xb5, 0x30, 0xc6, 0x69, 0xd8, 0x5b, 0x69, 0xd8, 0x5b, 0x45, 0x27, 0xea, 0x56, 0x39, 0xea, 0x46,
	0x19, 0xe8, 0x5a, 0xc6, 0xd6, 0x20, 0xb1, 0x11, 0xcc, 0x18, 0xdd, 0x58, 0x2d, 0x24, 0x09, 0xcd,
	0x18, 0x3d, 0x4d, 0x2d, 0x90, 0x3e, 0x47, 0xdd, 0xa8, 0x44, 0xdd, 0x2a, 0x88, 0x3e, 0x04, 0x1b,
	0xc1, 0x04, 0xeb, 0x43, 0x48, 0x68, 0x6a, 0xd8, 0x5b, 0x59, 0x45, 0x83, 0x58, 0xd5, 0x86, 0x3a,
	0xa6, 0xb5, 0x9a, 0xf6, 0x2a, 0x35, 0x71, 0x2a, 0x41, 0xbc, 0x10, 0x67, 0x3a, 0x42, 0x5c, 0x29,
	0x74, 0x91, 0x91, 0x4b, 0x08, 0xea, 0xf0, 0x92, 0x9c, 0x62, 0xb8, 0xe3, 0x99, 0xc0, 0x8e, 0x5f,
	0x34, 0xc8, 0x1c, 0xff, 0xe8, 0x9e, 0xbc, 0x40, 0x00, 0xd8, 0xb1, 0x21, 0x3e, 0xc1, 0x50, 0x1f,
	0x04, 0x09, 0x40, 0x9c, 0x1f, 0x92, 0x53, 0x0c, 0xf3, 0x93, 0x14, 0x17, 0x47, 0x69, 0x71, 0x6a,
	0x51, 0x5e, 0x62, 0x6e, 0xaa, 0x04, 0x33, 0xc8, 0x23, 0x41, 0x70, 0xbe, 0x53, 0xe2, 0x8d, 0x87,
	0x72, 0x0c, 0x3f, 0x1e, 0xca, 0x31, 0x36, 0x3c, 0x92, 0x63, 0x5c, 0xf1, 0x48, 0x8e, 0xf1, 0xc4,
	0x23, 0x39, 0xc6, 0x0b, 0x8f, 0xe4, 0x18, 0x1f, 0x3c, 0x92, 0x63, 0x7c, 0xf1, 0x48, 0x8e, 0xe1,
	0xc3, 0x23, 0x39, 0xc6, 0x09, 0x8f, 0xe5, 0x18, 0x16, 0x3c, 0x96, 0x63, 0x38, 0xf0, 0x58, 0x8e,
	0x31, 0x4a, 0x3f, 0x3d, 0x5f, 0xaf, 0x24, 0x23, 0xb5, 0x24, 0x23, 0x33, 0x2f, 0xbd, 0x58, 0x2f,
	0x2f, 0xb5, 0xa4, 0x3c, 0xbf, 0x28, 0x5b, 0x1f, 0x35, 0x99, 0x96, 0x19, 0xeb, 0x17, 0x64, 0xa7,
	0xeb, 0x97, 0x94, 0xe4, 0x15, 0x24, 0x25, 0xb1, 0x81, 0x53, 0x8a, 0x31, 0x20, 0x00, 0x00, 0xff,
	0xff, 0xfa, 0x9e, 0x3a, 0xf2, 0xcb, 0x02, 0x00, 0x00,
}

func (this *MQTTConnectionInfo) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*MQTTConnectionInfo)
	if !ok {
		that2, ok := that.(MQTTConnectionInfo)
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
	if this.PublicAddress != that1.PublicAddress {
		return false
	}
	if this.PublicTLSAddress != that1.PublicTLSAddress {
		return false
	}
	if this.Username != that1.Username {
		return false
	}
	return true
}
func (m *MQTTConnectionInfo) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *MQTTConnectionInfo) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *MQTTConnectionInfo) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if len(m.Username) > 0 {
		i -= len(m.Username)
		copy(dAtA[i:], m.Username)
		i = encodeVarintMqtt(dAtA, i, uint64(len(m.Username)))
		i--
		dAtA[i] = 0x1a
	}
	if len(m.PublicTLSAddress) > 0 {
		i -= len(m.PublicTLSAddress)
		copy(dAtA[i:], m.PublicTLSAddress)
		i = encodeVarintMqtt(dAtA, i, uint64(len(m.PublicTLSAddress)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.PublicAddress) > 0 {
		i -= len(m.PublicAddress)
		copy(dAtA[i:], m.PublicAddress)
		i = encodeVarintMqtt(dAtA, i, uint64(len(m.PublicAddress)))
		i--
		dAtA[i] = 0xa
	}
	return len(dAtA) - i, nil
}

func encodeVarintMqtt(dAtA []byte, offset int, v uint64) int {
	offset -= sovMqtt(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func NewPopulatedMQTTConnectionInfo(r randyMqtt, easy bool) *MQTTConnectionInfo {
	this := &MQTTConnectionInfo{}
	this.PublicAddress = string(randStringMqtt(r))
	this.PublicTLSAddress = string(randStringMqtt(r))
	this.Username = string(randStringMqtt(r))
	if !easy && r.Intn(10) != 0 {
	}
	return this
}

type randyMqtt interface {
	Float32() float32
	Float64() float64
	Int63() int64
	Int31() int32
	Uint32() uint32
	Intn(n int) int
}

func randUTF8RuneMqtt(r randyMqtt) rune {
	ru := r.Intn(62)
	if ru < 10 {
		return rune(ru + 48)
	} else if ru < 36 {
		return rune(ru + 55)
	}
	return rune(ru + 61)
}
func randStringMqtt(r randyMqtt) string {
	v1 := r.Intn(100)
	tmps := make([]rune, v1)
	for i := 0; i < v1; i++ {
		tmps[i] = randUTF8RuneMqtt(r)
	}
	return string(tmps)
}
func randUnrecognizedMqtt(r randyMqtt, maxFieldNumber int) (dAtA []byte) {
	l := r.Intn(5)
	for i := 0; i < l; i++ {
		wire := r.Intn(4)
		if wire == 3 {
			wire = 5
		}
		fieldNumber := maxFieldNumber + r.Intn(100)
		dAtA = randFieldMqtt(dAtA, r, fieldNumber, wire)
	}
	return dAtA
}
func randFieldMqtt(dAtA []byte, r randyMqtt, fieldNumber int, wire int) []byte {
	key := uint32(fieldNumber)<<3 | uint32(wire)
	switch wire {
	case 0:
		dAtA = encodeVarintPopulateMqtt(dAtA, uint64(key))
		v2 := r.Int63()
		if r.Intn(2) == 0 {
			v2 *= -1
		}
		dAtA = encodeVarintPopulateMqtt(dAtA, uint64(v2))
	case 1:
		dAtA = encodeVarintPopulateMqtt(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	case 2:
		dAtA = encodeVarintPopulateMqtt(dAtA, uint64(key))
		ll := r.Intn(100)
		dAtA = encodeVarintPopulateMqtt(dAtA, uint64(ll))
		for j := 0; j < ll; j++ {
			dAtA = append(dAtA, byte(r.Intn(256)))
		}
	default:
		dAtA = encodeVarintPopulateMqtt(dAtA, uint64(key))
		dAtA = append(dAtA, byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)), byte(r.Intn(256)))
	}
	return dAtA
}
func encodeVarintPopulateMqtt(dAtA []byte, v uint64) []byte {
	for v >= 1<<7 {
		dAtA = append(dAtA, uint8(uint64(v)&0x7f|0x80))
		v >>= 7
	}
	dAtA = append(dAtA, uint8(v))
	return dAtA
}
func (m *MQTTConnectionInfo) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	l = len(m.PublicAddress)
	if l > 0 {
		n += 1 + l + sovMqtt(uint64(l))
	}
	l = len(m.PublicTLSAddress)
	if l > 0 {
		n += 1 + l + sovMqtt(uint64(l))
	}
	l = len(m.Username)
	if l > 0 {
		n += 1 + l + sovMqtt(uint64(l))
	}
	return n
}

func sovMqtt(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozMqtt(x uint64) (n int) {
	return sovMqtt(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (this *MQTTConnectionInfo) String() string {
	if this == nil {
		return "nil"
	}
	s := strings.Join([]string{`&MQTTConnectionInfo{`,
		`PublicAddress:` + fmt.Sprintf("%v", this.PublicAddress) + `,`,
		`PublicTLSAddress:` + fmt.Sprintf("%v", this.PublicTLSAddress) + `,`,
		`Username:` + fmt.Sprintf("%v", this.Username) + `,`,
		`}`,
	}, "")
	return s
}
func valueToStringMqtt(v interface{}) string {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return "nil"
	}
	pv := reflect.Indirect(rv).Interface()
	return fmt.Sprintf("*%v", pv)
}
func (m *MQTTConnectionInfo) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowMqtt
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
			return fmt.Errorf("proto: MQTTConnectionInfo: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: MQTTConnectionInfo: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PublicAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMqtt
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
				return ErrInvalidLengthMqtt
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMqtt
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PublicAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field PublicTLSAddress", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMqtt
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
				return ErrInvalidLengthMqtt
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMqtt
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.PublicTLSAddress = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Username", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowMqtt
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
				return ErrInvalidLengthMqtt
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthMqtt
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Username = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipMqtt(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthMqtt
			}
			if (iNdEx + skippy) < 0 {
				return ErrInvalidLengthMqtt
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
func skipMqtt(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowMqtt
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
					return 0, ErrIntOverflowMqtt
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
					return 0, ErrIntOverflowMqtt
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
				return 0, ErrInvalidLengthMqtt
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupMqtt
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthMqtt
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthMqtt        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowMqtt          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupMqtt = fmt.Errorf("proto: unexpected end of group")
)
