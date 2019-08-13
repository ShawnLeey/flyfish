// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: backup.proto

package proto

import (
	fmt "fmt"
	github_com_golang_protobuf_proto "github.com/golang/protobuf/proto"
	proto "github.com/golang/protobuf/proto"
	io "io"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type SqlType int32

const (
	SqlType_insert SqlType = 1
	SqlType_update SqlType = 2
	SqlType_delete SqlType = 3
)

var SqlType_name = map[int32]string{
	1: "insert",
	2: "update",
	3: "delete",
}

var SqlType_value = map[string]int32{
	"insert": 1,
	"update": 2,
	"delete": 3,
}

func (x SqlType) Enum() *SqlType {
	p := new(SqlType)
	*p = x
	return p
}

func (x SqlType) String() string {
	return proto.EnumName(SqlType_name, int32(x))
}

func (x *SqlType) UnmarshalJSON(data []byte) error {
	value, err := proto.UnmarshalJSONEnum(SqlType_value, data, "SqlType")
	if err != nil {
		return err
	}
	*x = SqlType(value)
	return nil
}

func (SqlType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_65240d19de191688, []int{0}
}

type Record struct {
	Type                 *SqlType `protobuf:"varint,1,req,name=type,enum=proto.SqlType" json:"type,omitempty"`
	Table                *string  `protobuf:"bytes,2,req,name=table" json:"table,omitempty"`
	Key                  *string  `protobuf:"bytes,3,req,name=key" json:"key,omitempty"`
	WritebackVersion     *int64   `protobuf:"varint,4,req,name=writebackVersion" json:"writebackVersion,omitempty"`
	Fields               []*Field `protobuf:"bytes,5,rep,name=fields" json:"fields,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Record) Reset()         { *m = Record{} }
func (m *Record) String() string { return proto.CompactTextString(m) }
func (*Record) ProtoMessage()    {}
func (*Record) Descriptor() ([]byte, []int) {
	return fileDescriptor_65240d19de191688, []int{0}
}
func (m *Record) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *Record) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_Record.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalTo(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *Record) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Record.Merge(m, src)
}
func (m *Record) XXX_Size() int {
	return m.Size()
}
func (m *Record) XXX_DiscardUnknown() {
	xxx_messageInfo_Record.DiscardUnknown(m)
}

var xxx_messageInfo_Record proto.InternalMessageInfo

func (m *Record) GetType() SqlType {
	if m != nil && m.Type != nil {
		return *m.Type
	}
	return SqlType_insert
}

func (m *Record) GetTable() string {
	if m != nil && m.Table != nil {
		return *m.Table
	}
	return ""
}

func (m *Record) GetKey() string {
	if m != nil && m.Key != nil {
		return *m.Key
	}
	return ""
}

func (m *Record) GetWritebackVersion() int64 {
	if m != nil && m.WritebackVersion != nil {
		return *m.WritebackVersion
	}
	return 0
}

func (m *Record) GetFields() []*Field {
	if m != nil {
		return m.Fields
	}
	return nil
}

func init() {
	proto.RegisterEnum("proto.SqlType", SqlType_name, SqlType_value)
	proto.RegisterType((*Record)(nil), "proto.record")
}

func init() { proto.RegisterFile("backup.proto", fileDescriptor_65240d19de191688) }

var fileDescriptor_65240d19de191688 = []byte{
	// 226 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x49, 0x4a, 0x4c, 0xce,
	0x2e, 0x2d, 0xd0, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x05, 0x53, 0x52, 0xdc, 0x60, 0x0a,
	0x22, 0xa6, 0xb4, 0x8c, 0x91, 0x8b, 0xad, 0x28, 0x35, 0x39, 0xbf, 0x28, 0x45, 0x48, 0x89, 0x8b,
	0xa5, 0xa4, 0xb2, 0x20, 0x55, 0x82, 0x51, 0x81, 0x49, 0x83, 0xcf, 0x88, 0x0f, 0xa2, 0x40, 0x2f,
	0xb8, 0x30, 0x27, 0xa4, 0xb2, 0x20, 0x35, 0x08, 0x2c, 0x27, 0x24, 0xc2, 0xc5, 0x5a, 0x92, 0x98,
	0x94, 0x93, 0x2a, 0xc1, 0xa4, 0xc0, 0xa4, 0xc1, 0x19, 0x04, 0xe1, 0x08, 0x09, 0x70, 0x31, 0x67,
	0xa7, 0x56, 0x4a, 0x30, 0x83, 0xc5, 0x40, 0x4c, 0x21, 0x2d, 0x2e, 0x81, 0xf2, 0xa2, 0xcc, 0x92,
	0x54, 0x90, 0xfd, 0x61, 0xa9, 0x45, 0xc5, 0x99, 0xf9, 0x79, 0x12, 0x2c, 0x0a, 0x4c, 0x1a, 0xcc,
	0x41, 0x18, 0xe2, 0x42, 0x2a, 0x5c, 0x6c, 0x69, 0x99, 0xa9, 0x39, 0x29, 0xc5, 0x12, 0xac, 0x0a,
	0xcc, 0x1a, 0xdc, 0x46, 0x3c, 0x50, 0x9b, 0xc1, 0x82, 0x41, 0x50, 0x39, 0x2d, 0x5d, 0x2e, 0x76,
	0xa8, 0x53, 0x84, 0xb8, 0xb8, 0xd8, 0x32, 0xf3, 0x8a, 0x53, 0x8b, 0x4a, 0x04, 0x18, 0x41, 0xec,
	0xd2, 0x82, 0x94, 0xc4, 0x92, 0x54, 0x01, 0x26, 0x10, 0x3b, 0x25, 0x35, 0x27, 0xb5, 0x24, 0x55,
	0x80, 0xd9, 0x49, 0xe0, 0xc4, 0x23, 0x39, 0xc6, 0x0b, 0x8f, 0xe4, 0x18, 0x1f, 0x3c, 0x92, 0x63,
	0x9c, 0xf1, 0x58, 0x8e, 0x01, 0x10, 0x00, 0x00, 0xff, 0xff, 0xab, 0x57, 0xe1, 0x2a, 0x0c, 0x01,
	0x00, 0x00,
}

func (m *Record) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalTo(dAtA)
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *Record) MarshalTo(dAtA []byte) (int, error) {
	var i int
	_ = i
	var l int
	_ = l
	if m.Type == nil {
		return 0, new(github_com_golang_protobuf_proto.RequiredNotSetError)
	} else {
		dAtA[i] = 0x8
		i++
		i = encodeVarintBackup(dAtA, i, uint64(*m.Type))
	}
	if m.Table == nil {
		return 0, new(github_com_golang_protobuf_proto.RequiredNotSetError)
	} else {
		dAtA[i] = 0x12
		i++
		i = encodeVarintBackup(dAtA, i, uint64(len(*m.Table)))
		i += copy(dAtA[i:], *m.Table)
	}
	if m.Key == nil {
		return 0, new(github_com_golang_protobuf_proto.RequiredNotSetError)
	} else {
		dAtA[i] = 0x1a
		i++
		i = encodeVarintBackup(dAtA, i, uint64(len(*m.Key)))
		i += copy(dAtA[i:], *m.Key)
	}
	if m.WritebackVersion == nil {
		return 0, new(github_com_golang_protobuf_proto.RequiredNotSetError)
	} else {
		dAtA[i] = 0x20
		i++
		i = encodeVarintBackup(dAtA, i, uint64(*m.WritebackVersion))
	}
	if len(m.Fields) > 0 {
		for _, msg := range m.Fields {
			dAtA[i] = 0x2a
			i++
			i = encodeVarintBackup(dAtA, i, uint64(msg.Size()))
			n, err := msg.MarshalTo(dAtA[i:])
			if err != nil {
				return 0, err
			}
			i += n
		}
	}
	if m.XXX_unrecognized != nil {
		i += copy(dAtA[i:], m.XXX_unrecognized)
	}
	return i, nil
}

func encodeVarintBackup(dAtA []byte, offset int, v uint64) int {
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return offset + 1
}
func (m *Record) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if m.Type != nil {
		n += 1 + sovBackup(uint64(*m.Type))
	}
	if m.Table != nil {
		l = len(*m.Table)
		n += 1 + l + sovBackup(uint64(l))
	}
	if m.Key != nil {
		l = len(*m.Key)
		n += 1 + l + sovBackup(uint64(l))
	}
	if m.WritebackVersion != nil {
		n += 1 + sovBackup(uint64(*m.WritebackVersion))
	}
	if len(m.Fields) > 0 {
		for _, e := range m.Fields {
			l = e.Size()
			n += 1 + l + sovBackup(uint64(l))
		}
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovBackup(x uint64) (n int) {
	for {
		n++
		x >>= 7
		if x == 0 {
			break
		}
	}
	return n
}
func sozBackup(x uint64) (n int) {
	return sovBackup(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *Record) Unmarshal(dAtA []byte) error {
	var hasFields [1]uint64
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowBackup
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: record: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: record: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field Type", wireType)
			}
			var v SqlType
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBackup
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= (SqlType(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.Type = &v
			hasFields[0] |= uint64(0x00000001)
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Table", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBackup
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthBackup
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			s := string(dAtA[iNdEx:postIndex])
			m.Table = &s
			iNdEx = postIndex
			hasFields[0] |= uint64(0x00000002)
		case 3:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Key", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBackup
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= (uint64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthBackup
			}
			postIndex := iNdEx + intStringLen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			s := string(dAtA[iNdEx:postIndex])
			m.Key = &s
			iNdEx = postIndex
			hasFields[0] |= uint64(0x00000004)
		case 4:
			if wireType != 0 {
				return fmt.Errorf("proto: wrong wireType = %d for field WritebackVersion", wireType)
			}
			var v int64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBackup
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				v |= (int64(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			m.WritebackVersion = &v
			hasFields[0] |= uint64(0x00000008)
		case 5:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Fields", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowBackup
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthBackup
			}
			postIndex := iNdEx + msglen
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Fields = append(m.Fields, &Field{})
			if err := m.Fields[len(m.Fields)-1].Unmarshal(dAtA[iNdEx:postIndex]); err != nil {
				return err
			}
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipBackup(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if skippy < 0 {
				return ErrInvalidLengthBackup
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}
	if hasFields[0]&uint64(0x00000001) == 0 {
		return new(github_com_golang_protobuf_proto.RequiredNotSetError)
	}
	if hasFields[0]&uint64(0x00000002) == 0 {
		return new(github_com_golang_protobuf_proto.RequiredNotSetError)
	}
	if hasFields[0]&uint64(0x00000004) == 0 {
		return new(github_com_golang_protobuf_proto.RequiredNotSetError)
	}
	if hasFields[0]&uint64(0x00000008) == 0 {
		return new(github_com_golang_protobuf_proto.RequiredNotSetError)
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipBackup(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowBackup
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
					return 0, ErrIntOverflowBackup
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
			return iNdEx, nil
		case 1:
			iNdEx += 8
			return iNdEx, nil
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowBackup
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
			iNdEx += length
			if length < 0 {
				return 0, ErrInvalidLengthBackup
			}
			return iNdEx, nil
		case 3:
			for {
				var innerWire uint64
				var start int = iNdEx
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return 0, ErrIntOverflowBackup
					}
					if iNdEx >= l {
						return 0, io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					innerWire |= (uint64(b) & 0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				innerWireType := int(innerWire & 0x7)
				if innerWireType == 4 {
					break
				}
				next, err := skipBackup(dAtA[start:])
				if err != nil {
					return 0, err
				}
				iNdEx = start + next
			}
			return iNdEx, nil
		case 4:
			return iNdEx, nil
		case 5:
			iNdEx += 4
			return iNdEx, nil
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
	}
	panic("unreachable")
}

var (
	ErrInvalidLengthBackup = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowBackup   = fmt.Errorf("proto: integer overflow")
)
