// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        (unknown)
// source: filter/v1/filter_service.proto

package filterv1

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

var File_filter_v1_filter_service_proto protoreflect.FileDescriptor

const file_filter_v1_filter_service_proto_rawDesc = "" +
	"\n" +
	"\x1efilter/v1/filter_service.proto\x12\tfilter.v1\x1a\x1ffilter/v1/filter_messages.proto2W\n" +
	"\rFilterService\x12F\n" +
	"\tSetFilter\x12\x1b.filter.v1.SetFilterRequest\x1a\x1c.filter.v1.SetFilterResponseBEZCgithub.com/darvenommm/dating-bot-service/pkg/api/filter/v1;filterv1b\x06proto3"

var file_filter_v1_filter_service_proto_goTypes = []any{
	(*SetFilterRequest)(nil),  // 0: filter.v1.SetFilterRequest
	(*SetFilterResponse)(nil), // 1: filter.v1.SetFilterResponse
}
var file_filter_v1_filter_service_proto_depIdxs = []int32{
	0, // 0: filter.v1.FilterService.SetFilter:input_type -> filter.v1.SetFilterRequest
	1, // 1: filter.v1.FilterService.SetFilter:output_type -> filter.v1.SetFilterResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_filter_v1_filter_service_proto_init() }
func file_filter_v1_filter_service_proto_init() {
	if File_filter_v1_filter_service_proto != nil {
		return
	}
	file_filter_v1_filter_messages_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_filter_v1_filter_service_proto_rawDesc), len(file_filter_v1_filter_service_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_filter_v1_filter_service_proto_goTypes,
		DependencyIndexes: file_filter_v1_filter_service_proto_depIdxs,
	}.Build()
	File_filter_v1_filter_service_proto = out.File
	file_filter_v1_filter_service_proto_goTypes = nil
	file_filter_v1_filter_service_proto_depIdxs = nil
}
