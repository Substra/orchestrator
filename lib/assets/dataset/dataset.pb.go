// Copyright 2021 Owkin Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.14.0
// source: lib/assets/dataset/dataset.proto

package dataset

import (
	proto "github.com/golang/protobuf/proto"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

// Dataset references several related samples
type Dataset struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key        string            `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	SampleKeys []string          `protobuf:"bytes,2,rep,name=sample_keys,json=sampleKeys,proto3" json:"sample_keys,omitempty"`
	Worker     string            `protobuf:"bytes,3,opt,name=worker,proto3" json:"worker,omitempty"`
	Metadata   map[string]string `protobuf:"bytes,4,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Dataset) Reset() {
	*x = Dataset{}
	if protoimpl.UnsafeEnabled {
		mi := &file_lib_assets_dataset_dataset_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Dataset) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Dataset) ProtoMessage() {}

func (x *Dataset) ProtoReflect() protoreflect.Message {
	mi := &file_lib_assets_dataset_dataset_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Dataset.ProtoReflect.Descriptor instead.
func (*Dataset) Descriptor() ([]byte, []int) {
	return file_lib_assets_dataset_dataset_proto_rawDescGZIP(), []int{0}
}

func (x *Dataset) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *Dataset) GetSampleKeys() []string {
	if x != nil {
		return x.SampleKeys
	}
	return nil
}

func (x *Dataset) GetWorker() string {
	if x != nil {
		return x.Worker
	}
	return ""
}

func (x *Dataset) GetMetadata() map[string]string {
	if x != nil {
		return x.Metadata
	}
	return nil
}

var File_lib_assets_dataset_dataset_proto protoreflect.FileDescriptor

var file_lib_assets_dataset_dataset_proto_rawDesc = []byte{
	0x0a, 0x20, 0x6c, 0x69, 0x62, 0x2f, 0x61, 0x73, 0x73, 0x65, 0x74, 0x73, 0x2f, 0x64, 0x61, 0x74,
	0x61, 0x73, 0x65, 0x74, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x73, 0x65, 0x74, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x07, 0x64, 0x61, 0x74, 0x61, 0x73, 0x65, 0x74, 0x22, 0xcd, 0x01, 0x0a, 0x07,
	0x44, 0x61, 0x74, 0x61, 0x73, 0x65, 0x74, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x1f, 0x0a, 0x0b, 0x73, 0x61, 0x6d,
	0x70, 0x6c, 0x65, 0x5f, 0x6b, 0x65, 0x79, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x09, 0x52, 0x0a,
	0x73, 0x61, 0x6d, 0x70, 0x6c, 0x65, 0x4b, 0x65, 0x79, 0x73, 0x12, 0x16, 0x0a, 0x06, 0x77, 0x6f,
	0x72, 0x6b, 0x65, 0x72, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x77, 0x6f, 0x72, 0x6b,
	0x65, 0x72, 0x12, 0x3a, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x04,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x1e, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x73, 0x65, 0x74, 0x2e, 0x44,
	0x61, 0x74, 0x61, 0x73, 0x65, 0x74, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x45,
	0x6e, 0x74, 0x72, 0x79, 0x52, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x1a, 0x3b,
	0x0a, 0x0d, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12,
	0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65,
	0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x42, 0x32, 0x5a, 0x30, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6f, 0x77, 0x6b, 0x69, 0x6e, 0x2f,
	0x6f, 0x72, 0x63, 0x68, 0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x2f, 0x6c, 0x69, 0x62,
	0x2f, 0x61, 0x73, 0x73, 0x65, 0x74, 0x73, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x73, 0x65, 0x74, 0x62,
	0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_lib_assets_dataset_dataset_proto_rawDescOnce sync.Once
	file_lib_assets_dataset_dataset_proto_rawDescData = file_lib_assets_dataset_dataset_proto_rawDesc
)

func file_lib_assets_dataset_dataset_proto_rawDescGZIP() []byte {
	file_lib_assets_dataset_dataset_proto_rawDescOnce.Do(func() {
		file_lib_assets_dataset_dataset_proto_rawDescData = protoimpl.X.CompressGZIP(file_lib_assets_dataset_dataset_proto_rawDescData)
	})
	return file_lib_assets_dataset_dataset_proto_rawDescData
}

var file_lib_assets_dataset_dataset_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_lib_assets_dataset_dataset_proto_goTypes = []interface{}{
	(*Dataset)(nil), // 0: dataset.Dataset
	nil,             // 1: dataset.Dataset.MetadataEntry
}
var file_lib_assets_dataset_dataset_proto_depIdxs = []int32{
	1, // 0: dataset.Dataset.metadata:type_name -> dataset.Dataset.MetadataEntry
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_lib_assets_dataset_dataset_proto_init() }
func file_lib_assets_dataset_dataset_proto_init() {
	if File_lib_assets_dataset_dataset_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_lib_assets_dataset_dataset_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Dataset); i {
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
			RawDescriptor: file_lib_assets_dataset_dataset_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_lib_assets_dataset_dataset_proto_goTypes,
		DependencyIndexes: file_lib_assets_dataset_dataset_proto_depIdxs,
		MessageInfos:      file_lib_assets_dataset_dataset_proto_msgTypes,
	}.Build()
	File_lib_assets_dataset_dataset_proto = out.File
	file_lib_assets_dataset_dataset_proto_rawDesc = nil
	file_lib_assets_dataset_dataset_proto_goTypes = nil
	file_lib_assets_dataset_dataset_proto_depIdxs = nil
}
