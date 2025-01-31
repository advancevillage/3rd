// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.4
// 	protoc        v5.29.3
// source: dbx/sql.proto

package dbx

import (
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"

	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// 行数据
type SqlRow struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Column        [][]byte               `protobuf:"bytes,1,rep,name=column,proto3" json:"column,omitempty"` //列
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SqlRow) Reset() {
	*x = SqlRow{}
	mi := &file_dbx_sql_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SqlRow) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SqlRow) ProtoMessage() {}

func (x *SqlRow) ProtoReflect() protoreflect.Message {
	mi := &file_dbx_sql_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SqlRow.ProtoReflect.Descriptor instead.
func (*SqlRow) Descriptor() ([]byte, []int) {
	return file_dbx_sql_proto_rawDescGZIP(), []int{0}
}

func (x *SqlRow) GetColumn() [][]byte {
	if x != nil {
		return x.Column
	}
	return nil
}

type SqlReply struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Rows          []*SqlRow              `protobuf:"bytes,1,rep,name=rows,proto3" json:"rows,omitempty"`
	InsertId      int64                  `protobuf:"varint,2,opt,name=insertId,proto3" json:"insertId,omitempty"`         // 对于insert操作如果涉及到自增字段，可通过insert_id返回
	AffectedRows  int64                  `protobuf:"varint,3,opt,name=affectedRows,proto3" json:"affectedRows,omitempty"` // 返回受影响的行数
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SqlReply) Reset() {
	*x = SqlReply{}
	mi := &file_dbx_sql_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SqlReply) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SqlReply) ProtoMessage() {}

func (x *SqlReply) ProtoReflect() protoreflect.Message {
	mi := &file_dbx_sql_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SqlReply.ProtoReflect.Descriptor instead.
func (*SqlReply) Descriptor() ([]byte, []int) {
	return file_dbx_sql_proto_rawDescGZIP(), []int{1}
}

func (x *SqlReply) GetRows() []*SqlRow {
	if x != nil {
		return x.Rows
	}
	return nil
}

func (x *SqlReply) GetInsertId() int64 {
	if x != nil {
		return x.InsertId
	}
	return 0
}

func (x *SqlReply) GetAffectedRows() int64 {
	if x != nil {
		return x.AffectedRows
	}
	return 0
}

var File_dbx_sql_proto protoreflect.FileDescriptor

var file_dbx_sql_proto_rawDesc = string([]byte{
	0x0a, 0x0d, 0x64, 0x62, 0x78, 0x2f, 0x73, 0x71, 0x6c, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22,
	0x20, 0x0a, 0x06, 0x53, 0x71, 0x6c, 0x52, 0x6f, 0x77, 0x12, 0x16, 0x0a, 0x06, 0x63, 0x6f, 0x6c,
	0x75, 0x6d, 0x6e, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x06, 0x63, 0x6f, 0x6c, 0x75, 0x6d,
	0x6e, 0x22, 0x67, 0x0a, 0x08, 0x53, 0x71, 0x6c, 0x52, 0x65, 0x70, 0x6c, 0x79, 0x12, 0x1b, 0x0a,
	0x04, 0x72, 0x6f, 0x77, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x07, 0x2e, 0x53, 0x71,
	0x6c, 0x52, 0x6f, 0x77, 0x52, 0x04, 0x72, 0x6f, 0x77, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x69, 0x6e,
	0x73, 0x65, 0x72, 0x74, 0x49, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x69, 0x6e,
	0x73, 0x65, 0x72, 0x74, 0x49, 0x64, 0x12, 0x22, 0x0a, 0x0c, 0x61, 0x66, 0x66, 0x65, 0x63, 0x74,
	0x65, 0x64, 0x52, 0x6f, 0x77, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0c, 0x61, 0x66,
	0x66, 0x65, 0x63, 0x74, 0x65, 0x64, 0x52, 0x6f, 0x77, 0x73, 0x42, 0x27, 0x5a, 0x25, 0x67, 0x69,
	0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x61, 0x64, 0x76, 0x61, 0x6e, 0x63, 0x65,
	0x76, 0x69, 0x6c, 0x6c, 0x61, 0x67, 0x65, 0x2f, 0x33, 0x72, 0x64, 0x2f, 0x64, 0x62, 0x78, 0x3b,
	0x64, 0x62, 0x78, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
})

var (
	file_dbx_sql_proto_rawDescOnce sync.Once
	file_dbx_sql_proto_rawDescData []byte
)

func file_dbx_sql_proto_rawDescGZIP() []byte {
	file_dbx_sql_proto_rawDescOnce.Do(func() {
		file_dbx_sql_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_dbx_sql_proto_rawDesc), len(file_dbx_sql_proto_rawDesc)))
	})
	return file_dbx_sql_proto_rawDescData
}

var file_dbx_sql_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_dbx_sql_proto_goTypes = []any{
	(*SqlRow)(nil),   // 0: SqlRow
	(*SqlReply)(nil), // 1: SqlReply
}
var file_dbx_sql_proto_depIdxs = []int32{
	0, // 0: SqlReply.rows:type_name -> SqlRow
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_dbx_sql_proto_init() }
func file_dbx_sql_proto_init() {
	if File_dbx_sql_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_dbx_sql_proto_rawDesc), len(file_dbx_sql_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_dbx_sql_proto_goTypes,
		DependencyIndexes: file_dbx_sql_proto_depIdxs,
		MessageInfos:      file_dbx_sql_proto_msgTypes,
	}.Build()
	File_dbx_sql_proto = out.File
	file_dbx_sql_proto_goTypes = nil
	file_dbx_sql_proto_depIdxs = nil
}
