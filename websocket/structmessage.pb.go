// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.27.1
// 	protoc        v3.21.5
// source: websocket/structmessage.proto

package websocket

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type Format int32

const (
	Format_Json     Format = 0
	Format_Protobuf Format = 1
)

// Enum value maps for Format.
var (
	Format_name = map[int32]string{
		0: "Json",
		1: "Protobuf",
	}
	Format_value = map[string]int32{
		"Json":     0,
		"Protobuf": 1,
	}
)

func (x Format) Enum() *Format {
	p := new(Format)
	*p = x
	return p
}

func (x Format) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Format) Descriptor() protoreflect.EnumDescriptor {
	return file_websocket_structmessage_proto_enumTypes[0].Descriptor()
}

func (Format) Type() protoreflect.EnumType {
	return &file_websocket_structmessage_proto_enumTypes[0]
}

func (x Format) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Format.Descriptor instead.
func (Format) EnumDescriptor() ([]byte, []int) {
	return file_websocket_structmessage_proto_rawDescGZIP(), []int{0}
}

// StructMessage is a Protobuf based message on top of Message.
// It implements Message interface and can be sent with via a Speaker.
type StructMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Namespace string `protobuf:"bytes,1,opt,name=Namespace,proto3" json:"Namespace,omitempty"`
	Type      string `protobuf:"bytes,2,opt,name=Type,proto3" json:"Type,omitempty"`
	UUID      string `protobuf:"bytes,3,opt,name=UUID,proto3" json:"UUID,omitempty"`
	Status    int64  `protobuf:"varint,4,opt,name=Status,proto3" json:"Status,omitempty"`
	Format    Format `protobuf:"varint,5,opt,name=Format,proto3,enum=websocket.Format" json:"Format,omitempty"`
	Obj       []byte `protobuf:"bytes,6,opt,name=Obj,proto3" json:"Obj,omitempty"`
}

func (x *StructMessage) Reset() {
	*x = StructMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_websocket_structmessage_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *StructMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*StructMessage) ProtoMessage() {}

func (x *StructMessage) ProtoReflect() protoreflect.Message {
	mi := &file_websocket_structmessage_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use StructMessage.ProtoReflect.Descriptor instead.
func (*StructMessage) Descriptor() ([]byte, []int) {
	return file_websocket_structmessage_proto_rawDescGZIP(), []int{0}
}

func (x *StructMessage) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *StructMessage) GetType() string {
	if x != nil {
		return x.Type
	}
	return ""
}

func (x *StructMessage) GetUUID() string {
	if x != nil {
		return x.UUID
	}
	return ""
}

func (x *StructMessage) GetStatus() int64 {
	if x != nil {
		return x.Status
	}
	return 0
}

func (x *StructMessage) GetFormat() Format {
	if x != nil {
		return x.Format
	}
	return Format_Json
}

func (x *StructMessage) GetObj() []byte {
	if x != nil {
		return x.Obj
	}
	return nil
}

var File_websocket_structmessage_proto protoreflect.FileDescriptor

var file_websocket_structmessage_proto_rawDesc = []byte{
	0x0a, 0x1d, 0x77, 0x65, 0x62, 0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x2f, 0x73, 0x74, 0x72, 0x75,
	0x63, 0x74, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12,
	0x09, 0x77, 0x65, 0x62, 0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x22, 0xaa, 0x01, 0x0a, 0x0d, 0x53,
	0x74, 0x72, 0x75, 0x63, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x1c, 0x0a, 0x09,
	0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x09, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x54, 0x79,
	0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x54, 0x79, 0x70, 0x65, 0x12, 0x12,
	0x0a, 0x04, 0x55, 0x55, 0x49, 0x44, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x55, 0x55,
	0x49, 0x44, 0x12, 0x16, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x18, 0x04, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x29, 0x0a, 0x06, 0x46, 0x6f,
	0x72, 0x6d, 0x61, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x11, 0x2e, 0x77, 0x65, 0x62,
	0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x2e, 0x46, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x52, 0x06, 0x46,
	0x6f, 0x72, 0x6d, 0x61, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x4f, 0x62, 0x6a, 0x18, 0x06, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x03, 0x4f, 0x62, 0x6a, 0x2a, 0x20, 0x0a, 0x06, 0x46, 0x6f, 0x72, 0x6d, 0x61,
	0x74, 0x12, 0x08, 0x0a, 0x04, 0x4a, 0x73, 0x6f, 0x6e, 0x10, 0x00, 0x12, 0x0c, 0x0a, 0x08, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x10, 0x01, 0x42, 0x18, 0x5a, 0x16, 0x77, 0x65, 0x62,
	0x73, 0x6f, 0x63, 0x6b, 0x65, 0x74, 0x2d, 0x67, 0x6f, 0x2f, 0x77, 0x65, 0x62, 0x73, 0x6f, 0x63,
	0x6b, 0x65, 0x74, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_websocket_structmessage_proto_rawDescOnce sync.Once
	file_websocket_structmessage_proto_rawDescData = file_websocket_structmessage_proto_rawDesc
)

func file_websocket_structmessage_proto_rawDescGZIP() []byte {
	file_websocket_structmessage_proto_rawDescOnce.Do(func() {
		file_websocket_structmessage_proto_rawDescData = protoimpl.X.CompressGZIP(file_websocket_structmessage_proto_rawDescData)
	})
	return file_websocket_structmessage_proto_rawDescData
}

var file_websocket_structmessage_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_websocket_structmessage_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_websocket_structmessage_proto_goTypes = []interface{}{
	(Format)(0),           // 0: websocket.Format
	(*StructMessage)(nil), // 1: websocket.StructMessage
}
var file_websocket_structmessage_proto_depIdxs = []int32{
	0, // 0: websocket.StructMessage.Format:type_name -> websocket.Format
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_websocket_structmessage_proto_init() }
func file_websocket_structmessage_proto_init() {
	if File_websocket_structmessage_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_websocket_structmessage_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*StructMessage); i {
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
			RawDescriptor: file_websocket_structmessage_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_websocket_structmessage_proto_goTypes,
		DependencyIndexes: file_websocket_structmessage_proto_depIdxs,
		EnumInfos:         file_websocket_structmessage_proto_enumTypes,
		MessageInfos:      file_websocket_structmessage_proto_msgTypes,
	}.Build()
	File_websocket_structmessage_proto = out.File
	file_websocket_structmessage_proto_rawDesc = nil
	file_websocket_structmessage_proto_goTypes = nil
	file_websocket_structmessage_proto_depIdxs = nil
}