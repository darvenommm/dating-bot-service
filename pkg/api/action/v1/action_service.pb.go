// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        (unknown)
// source: action/v1/action_service.proto

package actionv1

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

var File_action_v1_action_service_proto protoreflect.FileDescriptor

const file_action_v1_action_service_proto_rawDesc = "" +
	"\n" +
	"\x1eaction/v1/action_service.proto\x12\taction.v1\x1a\x1faction/v1/action_messages.proto2W\n" +
	"\rActionService\x12F\n" +
	"\tAddAction\x12\x1b.action.v1.AddActionRequest\x1a\x1c.action.v1.AddActionResponseBEZCgithub.com/darvenommm/dating-bot-service/pkg/api/action/v1;actionv1b\x06proto3"

var file_action_v1_action_service_proto_goTypes = []any{
	(*AddActionRequest)(nil),  // 0: action.v1.AddActionRequest
	(*AddActionResponse)(nil), // 1: action.v1.AddActionResponse
}
var file_action_v1_action_service_proto_depIdxs = []int32{
	0, // 0: action.v1.ActionService.AddAction:input_type -> action.v1.AddActionRequest
	1, // 1: action.v1.ActionService.AddAction:output_type -> action.v1.AddActionResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_action_v1_action_service_proto_init() }
func file_action_v1_action_service_proto_init() {
	if File_action_v1_action_service_proto != nil {
		return
	}
	file_action_v1_action_messages_proto_init()
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_action_v1_action_service_proto_rawDesc), len(file_action_v1_action_service_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   0,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_action_v1_action_service_proto_goTypes,
		DependencyIndexes: file_action_v1_action_service_proto_depIdxs,
	}.Build()
	File_action_v1_action_service_proto = out.File
	file_action_v1_action_service_proto_goTypes = nil
	file_action_v1_action_service_proto_depIdxs = nil
}
