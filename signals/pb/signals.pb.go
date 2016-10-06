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
	Id          uint64          `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Org         string          `protobuf:"bytes,2,opt,name=org" json:"org,omitempty"`
	App         string          `protobuf:"bytes,3,opt,name=app" json:"app,omitempty"`
	Namespace   string          `protobuf:"bytes,4,opt,name=namespace" json:"namespace,omitempty"`
	Arrvied     string          `protobuf:"bytes,5,opt,name=arrvied" json:"arrvied,omitempty"`
	Interaction InteractionType `protobuf:"varint,6,opt,name=interaction,enum=pb.InteractionType" json:"interaction,omitempty"`
	Content     *Content        `protobuf:"bytes,7,opt,name=content" json:"content,omitempty"`
	User        *User           `protobuf:"bytes,8,opt,name=user" json:"user,omitempty"`
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
	// 540 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x5c, 0x53, 0xcd, 0x6a, 0xdb, 0x4c,
	0x14, 0x8d, 0x6c, 0x45, 0x3f, 0x57, 0x89, 0x33, 0xcc, 0x07, 0x1f, 0x5a, 0x64, 0x61, 0xdc, 0x06,
	0x4c, 0x16, 0x2e, 0xa4, 0x74, 0x5b, 0xd0, 0xcf, 0x84, 0x0c, 0xb1, 0x35, 0x66, 0x34, 0xb2, 0xbb,
	0x13, 0xb2, 0xad, 0x1a, 0x41, 0x22, 0x09, 0x49, 0x2e, 0xe4, 0x7d, 0xfa, 0x2a, 0x7d, 0x9f, 0x3e,
	0x42, 0x99, 0x91, 0x4d, 0x4c, 0x76, 0xd2, 0x39, 0xe7, 0x9e, 0x39, 0xf7, 0xc0, 0x85, 0xeb, 0xb6,
	0xd8, 0x97, 0xd9, 0x4b, 0x3b, 0xab, 0x9b, 0xaa, 0xab, 0xf0, 0xa0, 0xde, 0x4c, 0xbe, 0x83, 0x19,
	0x54, 0x65, 0x97, 0x97, 0x1d, 0x1e, 0xc1, 0xa0, 0xd8, 0xb9, 0xda, 0x58, 0x9b, 0xda, 0x7c, 0x50,
	0xec, 0xf0, 0x27, 0xd0, 0xbb, 0xb7, 0x3a, 0x77, 0x07, 0x63, 0x6d, 0x3a, 0x7a, 0xb8, 0x99, 0xd5,
	0x9b, 0xd9, 0x51, 0x2a, 0xde, 0xea, 0x9c, 0x2b, 0x72, 0xf2, 0x57, 0x03, 0x23, 0x56, 0xae, 0x67,
	0xf3, 0xba, 0x9a, 0x47, 0x30, 0xac, 0x9a, 0xbd, 0x1a, 0xb7, 0xb9, 0xfc, 0x94, 0x48, 0x56, 0xd7,
	0xee, 0xb0, 0x47, 0xb2, 0xba, 0xc6, 0xb7, 0x60, 0x97, 0xd9, 0x6b, 0xde, 0xd6, 0xd9, 0x36, 0x77,
	0x75, 0x85, 0xbf, 0x03, 0xd8, 0x05, 0x33, 0x6b, 0x9a, 0x5f, 0x45, 0xbe, 0x73, 0x2f, 0x15, 0x77,
	0xfa, 0xc5, 0xdf, 0xc0, 0x29, 0xca, 0x2e, 0x6f, 0xb2, 0x6d, 0x57, 0x54, 0xa5, 0x6b, 0xa8, 0x88,
	0xff, 0xc9, 0x88, 0xf4, 0x1d, 0x56, 0x31, 0xcf, 0x75, 0xf8, 0x0e, 0xcc, 0x6d, 0xbf, 0x82, 0x6b,
	0x8e, 0xb5, 0xa9, 0xf3, 0xe0, 0x9c, 0x6d, 0xc5, 0x4f, 0x1c, 0xbe, 0x05, 0xfd, 0xd0, 0xe6, 0x8d,
	0x6b, 0x29, 0x8d, 0x25, 0x35, 0x49, 0x9b, 0x37, 0x5c, 0xa1, 0x93, 0x2f, 0x60, 0xf6, 0x1b, 0xb7,
	0xf8, 0x33, 0x98, 0xc7, 0x4a, 0x5d, 0x6d, 0x3c, 0x9c, 0x3a, 0x0f, 0x20, 0xb5, 0x3d, 0xcb, 0x4f,
	0xd4, 0xe4, 0x7f, 0xd0, 0xe5, 0xf8, 0xc7, 0x82, 0xef, 0x7f, 0x6b, 0xe0, 0x9c, 0x35, 0x8a, 0x1d,
	0x30, 0x93, 0xe8, 0x39, 0x62, 0xeb, 0x08, 0x5d, 0xe0, 0x11, 0x80, 0x97, 0x84, 0x94, 0xa5, 0x3e,
	0x63, 0xcf, 0x48, 0xc3, 0x36, 0x5c, 0x2e, 0xd8, 0x8a, 0x12, 0x34, 0xc0, 0x08, 0xae, 0x1e, 0xbd,
	0xf8, 0x89, 0xb2, 0x28, 0xa5, 0x82, 0x2c, 0xd0, 0x10, 0x63, 0x18, 0x3d, 0x26, 0x3c, 0xa2, 0x22,
	0xe1, 0xa4, 0xc7, 0x74, 0xa9, 0x8a, 0xc8, 0x3a, 0x4e, 0x3d, 0x2e, 0x68, 0x30, 0x27, 0xe8, 0x12,
	0x5f, 0x83, 0x2d, 0xc8, 0x0f, 0xd1, 0x3b, 0x1a, 0x18, 0xc0, 0xe0, 0x24, 0xa0, 0x4b, 0x82, 0x4c,
	0xf9, 0x1a, 0x27, 0xb1, 0xf0, 0x12, 0xee, 0x45, 0x02, 0x59, 0x3d, 0xb7, 0xa2, 0x64, 0x8d, 0xec,
	0xfb, 0x3f, 0x1a, 0xdc, 0x7c, 0x68, 0x15, 0xdf, 0x80, 0xe3, 0x85, 0x61, 0x2a, 0x58, 0x1a, 0x78,
	0x5c, 0xa0, 0x0b, 0xe9, 0x2d, 0x6d, 0x17, 0xbe, 0xc7, 0x65, 0xda, 0x2b, 0xb0, 0x02, 0xb6, 0x58,
	0xce, 0x89, 0x20, 0x68, 0x28, 0x17, 0x0b, 0x69, 0x3c, 0xa7, 0xcf, 0x04, 0xe9, 0x52, 0x19, 0xb2,
	0x75, 0x94, 0xae, 0x98, 0x90, 0xa1, 0x00, 0x8c, 0x47, 0x36, 0x9f, 0xb3, 0x35, 0x32, 0xb0, 0x05,
	0xba, 0x12, 0x99, 0x72, 0x5b, 0xc6, 0x43, 0xc2, 0x91, 0x25, 0x41, 0xee, 0x09, 0x82, 0x6c, 0x09,
	0xc6, 0x4f, 0x1e, 0x27, 0x08, 0xa4, 0x49, 0x9c, 0xf8, 0x71, 0xc0, 0xa9, 0x4f, 0x90, 0x23, 0x35,
	0x2a, 0xec, 0x95, 0x7c, 0x38, 0x89, 0x8e, 0x86, 0xd7, 0xaa, 0xd1, 0x65, 0xff, 0xd2, 0xc8, 0xbf,
	0x03, 0x77, 0x5b, 0xbd, 0xce, 0xba, 0xac, 0xde, 0xbf, 0x1c, 0xf2, 0xd9, 0xe9, 0x16, 0x76, 0x59,
	0x97, 0xf9, 0xf6, 0x52, 0x5e, 0xc4, 0xe6, 0xf0, 0xb3, 0xdd, 0x18, 0xea, 0x38, 0xbe, 0xfe, 0x0b,
	0x00, 0x00, 0xff, 0xff, 0xc5, 0xa8, 0xb5, 0xa2, 0x2d, 0x03, 0x00, 0x00,
}
