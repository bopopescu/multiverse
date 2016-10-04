// Code generated by protoc-gen-go.
// source: signals.proto
// DO NOT EDIT!

/*
Package pb is a generated protocol buffer package.

It is generated from these files:
	signals.proto

It has these top-level messages:
	Content
	Signal
	Signals
	User
*/
package pb

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type ContentType int32

const (
	ContentType_UNKNOWN        ContentType = 0
	ContentType_AUDIO_BOOK     ContentType = 1
	ContentType_MOVIE          ContentType = 2
	ContentType_FASHION_ITEM   ContentType = 3
	ContentType_FURNITURE_ITEM ContentType = 4
	ContentType_NEWS_ARTICLE   ContentType = 5
	ContentType_TEXT_BOOK      ContentType = 6
	ContentType_RECIPE         ContentType = 7
	ContentType_RESTAURANT     ContentType = 8
	ContentType_REVIEW         ContentType = 9
)

var ContentType_name = map[int32]string{
	0: "UNKNOWN",
	1: "AUDIO_BOOK",
	2: "MOVIE",
	3: "FASHION_ITEM",
	4: "FURNITURE_ITEM",
	5: "NEWS_ARTICLE",
	6: "TEXT_BOOK",
	7: "RECIPE",
	8: "RESTAURANT",
	9: "REVIEW",
}
var ContentType_value = map[string]int32{
	"UNKNOWN":        0,
	"AUDIO_BOOK":     1,
	"MOVIE":          2,
	"FASHION_ITEM":   3,
	"FURNITURE_ITEM": 4,
	"NEWS_ARTICLE":   5,
	"TEXT_BOOK":      6,
	"RECIPE":         7,
	"RESTAURANT":     8,
	"REVIEW":         9,
}

func (x ContentType) String() string {
	return proto.EnumName(ContentType_name, int32(x))
}
func (ContentType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type InteractionType int32

const (
	InteractionType_ADD_TO_CART InteractionType = 0
	InteractionType_BOOKMBARK   InteractionType = 1
	InteractionType_COMPLETE    InteractionType = 3
	InteractionType_DISLIKE     InteractionType = 4
	InteractionType_DOWN_VOTE   InteractionType = 5
	InteractionType_FOLLOW      InteractionType = 6
	InteractionType_LIKE        InteractionType = 7
	InteractionType_ORDER       InteractionType = 8
	InteractionType_RATE        InteractionType = 9
	InteractionType_SHARE       InteractionType = 10
	InteractionType_SUBSCRIBE   InteractionType = 11
	InteractionType_VIEW        InteractionType = 12
	InteractionType_UNFOLLOW    InteractionType = 13
	InteractionType_UP_VOTE     InteractionType = 14
)

var InteractionType_name = map[int32]string{
	0:  "ADD_TO_CART",
	1:  "BOOKMBARK",
	3:  "COMPLETE",
	4:  "DISLIKE",
	5:  "DOWN_VOTE",
	6:  "FOLLOW",
	7:  "LIKE",
	8:  "ORDER",
	9:  "RATE",
	10: "SHARE",
	11: "SUBSCRIBE",
	12: "VIEW",
	13: "UNFOLLOW",
	14: "UP_VOTE",
}
var InteractionType_value = map[string]int32{
	"ADD_TO_CART": 0,
	"BOOKMBARK":   1,
	"COMPLETE":    3,
	"DISLIKE":     4,
	"DOWN_VOTE":   5,
	"FOLLOW":      6,
	"LIKE":        7,
	"ORDER":       8,
	"RATE":        9,
	"SHARE":       10,
	"SUBSCRIBE":   11,
	"VIEW":        12,
	"UNFOLLOW":    13,
	"UP_VOTE":     14,
}

func (x InteractionType) String() string {
	return proto.EnumName(InteractionType_name, int32(x))
}
func (InteractionType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type Content struct {
	Id   string      `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Type ContentType `protobuf:"varint,2,opt,name=type,enum=pb.ContentType" json:"type,omitempty"`
}

func (m *Content) Reset()                    { *m = Content{} }
func (m *Content) String() string            { return proto.CompactTextString(m) }
func (*Content) ProtoMessage()               {}
func (*Content) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type Signal struct {
	Id          string          `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Org         string          `protobuf:"bytes,2,opt,name=org" json:"org,omitempty"`
	App         string          `protobuf:"bytes,3,opt,name=app" json:"app,omitempty"`
	Arrvied     string          `protobuf:"bytes,4,opt,name=arrvied" json:"arrvied,omitempty"`
	Interaction InteractionType `protobuf:"varint,5,opt,name=interaction,enum=pb.InteractionType" json:"interaction,omitempty"`
	Content     *Content        `protobuf:"bytes,6,opt,name=content" json:"content,omitempty"`
	User        *User           `protobuf:"bytes,7,opt,name=user" json:"user,omitempty"`
}

func (m *Signal) Reset()                    { *m = Signal{} }
func (m *Signal) String() string            { return proto.CompactTextString(m) }
func (*Signal) ProtoMessage()               {}
func (*Signal) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Signal) GetContent() *Content {
	if m != nil {
		return m.Content
	}
	return nil
}

func (m *Signal) GetUser() *User {
	if m != nil {
		return m.User
	}
	return nil
}

type Signals struct {
	Signals []*Signal `protobuf:"bytes,1,rep,name=signals" json:"signals,omitempty"`
}

func (m *Signals) Reset()                    { *m = Signals{} }
func (m *Signals) String() string            { return proto.CompactTextString(m) }
func (*Signals) ProtoMessage()               {}
func (*Signals) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Signals) GetSignals() []*Signal {
	if m != nil {
		return m.Signals
	}
	return nil
}

type User struct {
	Id string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
}

func (m *User) Reset()                    { *m = User{} }
func (m *User) String() string            { return proto.CompactTextString(m) }
func (*User) ProtoMessage()               {}
func (*User) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func init() {
	proto.RegisterType((*Content)(nil), "pb.Content")
	proto.RegisterType((*Signal)(nil), "pb.Signal")
	proto.RegisterType((*Signals)(nil), "pb.Signals")
	proto.RegisterType((*User)(nil), "pb.User")
	proto.RegisterEnum("pb.ContentType", ContentType_name, ContentType_value)
	proto.RegisterEnum("pb.InteractionType", InteractionType_name, InteractionType_value)
}

func init() { proto.RegisterFile("signals.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 523 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x64, 0x52, 0xdf, 0x8a, 0xda, 0x4e,
	0x18, 0xdd, 0x68, 0x36, 0x31, 0x5f, 0x56, 0x1d, 0xe6, 0x07, 0x3f, 0x72, 0xd1, 0x0b, 0xb1, 0x5d,
	0x90, 0xbd, 0xb0, 0x60, 0xe9, 0x6d, 0x21, 0x7f, 0x46, 0x76, 0x50, 0x33, 0x32, 0x99, 0x68, 0xef,
	0x42, 0xd4, 0x54, 0x02, 0x5b, 0x13, 0x92, 0x58, 0xd8, 0xf7, 0xe9, 0xab, 0xf4, 0x05, 0xfa, 0x44,
	0x65, 0x26, 0x4a, 0x65, 0x7b, 0x97, 0x9c, 0x73, 0xe6, 0x7c, 0xe7, 0x3b, 0x7c, 0xd0, 0xaf, 0xf3,
	0xe3, 0x29, 0x7d, 0xa9, 0xa7, 0x65, 0x55, 0x34, 0x05, 0xee, 0x94, 0xbb, 0xf1, 0x17, 0x30, 0xfd,
	0xe2, 0xd4, 0x64, 0xa7, 0x06, 0x0f, 0xa0, 0x93, 0x1f, 0x1c, 0x6d, 0xa4, 0x4d, 0x2c, 0xde, 0xc9,
	0x0f, 0xf8, 0x3d, 0xe8, 0xcd, 0x6b, 0x99, 0x39, 0x9d, 0x91, 0x36, 0x19, 0xcc, 0x86, 0xd3, 0x72,
	0x37, 0xbd, 0x48, 0xc5, 0x6b, 0x99, 0x71, 0x45, 0x8e, 0x7f, 0x6b, 0x60, 0x44, 0xca, 0xf5, 0x9f,
	0xf7, 0x08, 0xba, 0x45, 0x75, 0x54, 0xcf, 0x2d, 0x2e, 0x3f, 0x25, 0x92, 0x96, 0xa5, 0xd3, 0x6d,
	0x91, 0xb4, 0x2c, 0xb1, 0x03, 0x66, 0x5a, 0x55, 0x3f, 0xf2, 0xec, 0xe0, 0xe8, 0x0a, 0xbd, 0xfe,
	0xe2, 0xcf, 0x60, 0xe7, 0xa7, 0x26, 0xab, 0xd2, 0x7d, 0x93, 0x17, 0x27, 0xe7, 0x5e, 0x85, 0xf8,
	0x4f, 0x86, 0xa0, 0x7f, 0x61, 0x15, 0xe4, 0x56, 0x87, 0x1f, 0xc1, 0xdc, 0xb7, 0x21, 0x1d, 0x63,
	0xa4, 0x4d, 0xec, 0x99, 0x7d, 0x93, 0x9b, 0x5f, 0x39, 0xfc, 0x0e, 0xf4, 0x73, 0x9d, 0x55, 0x8e,
	0xa9, 0x34, 0x3d, 0xa9, 0x89, 0xeb, 0xac, 0xe2, 0x0a, 0x1d, 0x7f, 0x04, 0xb3, 0xdd, 0xa9, 0xc6,
	0x1f, 0xc0, 0xbc, 0x94, 0xe6, 0x68, 0xa3, 0xee, 0xc4, 0x9e, 0x81, 0xd4, 0xb6, 0x2c, 0xbf, 0x52,
	0xe3, 0xff, 0x41, 0x97, 0xcf, 0xdf, 0x56, 0xf0, 0xf4, 0x53, 0x03, 0xfb, 0xa6, 0x33, 0x6c, 0x83,
	0x19, 0x87, 0x8b, 0x90, 0x6d, 0x43, 0x74, 0x87, 0x07, 0x00, 0x6e, 0x1c, 0x50, 0x96, 0x78, 0x8c,
	0x2d, 0x90, 0x86, 0x2d, 0xb8, 0x5f, 0xb1, 0x0d, 0x25, 0xa8, 0x83, 0x11, 0x3c, 0xcc, 0xdd, 0xe8,
	0x99, 0xb2, 0x30, 0xa1, 0x82, 0xac, 0x50, 0x17, 0x63, 0x18, 0xcc, 0x63, 0x1e, 0x52, 0x11, 0x73,
	0xd2, 0x62, 0xba, 0x54, 0x85, 0x64, 0x1b, 0x25, 0x2e, 0x17, 0xd4, 0x5f, 0x12, 0x74, 0x8f, 0xfb,
	0x60, 0x09, 0xf2, 0x55, 0xb4, 0x8e, 0x06, 0x06, 0x30, 0x38, 0xf1, 0xe9, 0x9a, 0x20, 0x53, 0x4e,
	0xe3, 0x24, 0x12, 0x6e, 0xcc, 0xdd, 0x50, 0xa0, 0x5e, 0xcb, 0x6d, 0x28, 0xd9, 0x22, 0xeb, 0xe9,
	0x97, 0x06, 0xc3, 0x37, 0xad, 0xe2, 0x21, 0xd8, 0x6e, 0x10, 0x24, 0x82, 0x25, 0xbe, 0xcb, 0x05,
	0xba, 0x93, 0xde, 0xd2, 0x76, 0xe5, 0xb9, 0x5c, 0xa6, 0x7d, 0x80, 0x9e, 0xcf, 0x56, 0xeb, 0x25,
	0x11, 0x04, 0x75, 0xe5, 0x62, 0x01, 0x8d, 0x96, 0x74, 0x41, 0x90, 0x2e, 0x95, 0x01, 0xdb, 0x86,
	0xc9, 0x86, 0x09, 0x19, 0x0a, 0xc0, 0x98, 0xb3, 0xe5, 0x92, 0x6d, 0x91, 0x81, 0x7b, 0xa0, 0x2b,
	0x91, 0x29, 0xb7, 0x65, 0x3c, 0x20, 0x1c, 0xf5, 0x24, 0xc8, 0x5d, 0x41, 0x90, 0x25, 0xc1, 0xe8,
	0xd9, 0xe5, 0x04, 0x81, 0x34, 0x89, 0x62, 0x2f, 0xf2, 0x39, 0xf5, 0x08, 0xb2, 0xa5, 0x46, 0x85,
	0x7d, 0x90, 0x83, 0xe3, 0xf0, 0x62, 0xd8, 0x57, 0x8d, 0xae, 0xdb, 0x49, 0x03, 0xef, 0x11, 0x9c,
	0x7d, 0xf1, 0x7d, 0xda, 0xa4, 0xe5, 0xf1, 0xe5, 0x9c, 0x4d, 0xaf, 0xd7, 0x7e, 0x48, 0x9b, 0xd4,
	0xb3, 0xd6, 0xf2, 0xe6, 0x77, 0xe7, 0x6f, 0xf5, 0xce, 0x50, 0xe7, 0xff, 0xe9, 0x4f, 0x00, 0x00,
	0x00, 0xff, 0xff, 0xe7, 0x7a, 0xb2, 0xb7, 0x0f, 0x03, 0x00, 0x00,
}
