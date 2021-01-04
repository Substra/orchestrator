// Copyright 2020 Owkin Inc.
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
// source: lib/assets/objective/objective.proto

package objective

import (
	proto "github.com/golang/protobuf/proto"
	assets "github.com/owkin/orchestrator/lib/assets"
	dataset "github.com/owkin/orchestrator/lib/assets/dataset"
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

// Objective represents the hypothesis against which a model is trained and evaluated.
type Objective struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key         string              `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Name        string              `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	TestDataset *dataset.Dataset    `protobuf:"bytes,3,opt,name=test_dataset,json=testDataset,proto3" json:"test_dataset,omitempty"`
	Description *assets.Addressable `protobuf:"bytes,4,opt,name=description,proto3" json:"description,omitempty"`
	MetricsName string              `protobuf:"bytes,5,opt,name=metrics_name,json=metricsName,proto3" json:"metrics_name,omitempty"`
	Metrics     *assets.Addressable `protobuf:"bytes,6,opt,name=metrics,proto3" json:"metrics,omitempty"`
	Permissions *assets.Permissions `protobuf:"bytes,7,opt,name=permissions,proto3" json:"permissions,omitempty"`
	Metadata    map[string]string   `protobuf:"bytes,8,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *Objective) Reset() {
	*x = Objective{}
	if protoimpl.UnsafeEnabled {
		mi := &file_lib_assets_objective_objective_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Objective) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Objective) ProtoMessage() {}

func (x *Objective) ProtoReflect() protoreflect.Message {
	mi := &file_lib_assets_objective_objective_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Objective.ProtoReflect.Descriptor instead.
func (*Objective) Descriptor() ([]byte, []int) {
	return file_lib_assets_objective_objective_proto_rawDescGZIP(), []int{0}
}

func (x *Objective) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *Objective) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Objective) GetTestDataset() *dataset.Dataset {
	if x != nil {
		return x.TestDataset
	}
	return nil
}

func (x *Objective) GetDescription() *assets.Addressable {
	if x != nil {
		return x.Description
	}
	return nil
}

func (x *Objective) GetMetricsName() string {
	if x != nil {
		return x.MetricsName
	}
	return ""
}

func (x *Objective) GetMetrics() *assets.Addressable {
	if x != nil {
		return x.Metrics
	}
	return nil
}

func (x *Objective) GetPermissions() *assets.Permissions {
	if x != nil {
		return x.Permissions
	}
	return nil
}

func (x *Objective) GetMetadata() map[string]string {
	if x != nil {
		return x.Metadata
	}
	return nil
}

type ObjectiveQueryParam struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ObjectiveQueryParam) Reset() {
	*x = ObjectiveQueryParam{}
	if protoimpl.UnsafeEnabled {
		mi := &file_lib_assets_objective_objective_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ObjectiveQueryParam) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ObjectiveQueryParam) ProtoMessage() {}

func (x *ObjectiveQueryParam) ProtoReflect() protoreflect.Message {
	mi := &file_lib_assets_objective_objective_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ObjectiveQueryParam.ProtoReflect.Descriptor instead.
func (*ObjectiveQueryParam) Descriptor() ([]byte, []int) {
	return file_lib_assets_objective_objective_proto_rawDescGZIP(), []int{1}
}

type ObjectiveQueryResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Objectives []*Objective `protobuf:"bytes,1,rep,name=objectives,proto3" json:"objectives,omitempty"`
}

func (x *ObjectiveQueryResponse) Reset() {
	*x = ObjectiveQueryResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_lib_assets_objective_objective_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ObjectiveQueryResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ObjectiveQueryResponse) ProtoMessage() {}

func (x *ObjectiveQueryResponse) ProtoReflect() protoreflect.Message {
	mi := &file_lib_assets_objective_objective_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ObjectiveQueryResponse.ProtoReflect.Descriptor instead.
func (*ObjectiveQueryResponse) Descriptor() ([]byte, []int) {
	return file_lib_assets_objective_objective_proto_rawDescGZIP(), []int{2}
}

func (x *ObjectiveQueryResponse) GetObjectives() []*Objective {
	if x != nil {
		return x.Objectives
	}
	return nil
}

type LeaderboardQueryParam struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	ObjectiveKey string           `protobuf:"bytes,1,opt,name=objective_key,json=objectiveKey,proto3" json:"objective_key,omitempty"`
	SortOrder    assets.SortOrder `protobuf:"varint,2,opt,name=sort_order,json=sortOrder,proto3,enum=assets.SortOrder" json:"sort_order,omitempty"`
}

func (x *LeaderboardQueryParam) Reset() {
	*x = LeaderboardQueryParam{}
	if protoimpl.UnsafeEnabled {
		mi := &file_lib_assets_objective_objective_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LeaderboardQueryParam) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LeaderboardQueryParam) ProtoMessage() {}

func (x *LeaderboardQueryParam) ProtoReflect() protoreflect.Message {
	mi := &file_lib_assets_objective_objective_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LeaderboardQueryParam.ProtoReflect.Descriptor instead.
func (*LeaderboardQueryParam) Descriptor() ([]byte, []int) {
	return file_lib_assets_objective_objective_proto_rawDescGZIP(), []int{3}
}

func (x *LeaderboardQueryParam) GetObjectiveKey() string {
	if x != nil {
		return x.ObjectiveKey
	}
	return ""
}

func (x *LeaderboardQueryParam) GetSortOrder() assets.SortOrder {
	if x != nil {
		return x.SortOrder
	}
	return assets.SortOrder_ASCENDING
}

// This will probably live somewhere else once we implement test tuple assets
type BoardTuple struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Algo          string  `protobuf:"bytes,1,opt,name=algo,proto3" json:"algo,omitempty"`
	Creator       string  `protobuf:"bytes,2,opt,name=creator,proto3" json:"creator,omitempty"`
	Key           string  `protobuf:"bytes,3,opt,name=key,proto3" json:"key,omitempty"`
	TraintupleKey string  `protobuf:"bytes,4,opt,name=traintuple_key,json=traintupleKey,proto3" json:"traintuple_key,omitempty"`
	Perf          float32 `protobuf:"fixed32,5,opt,name=perf,proto3" json:"perf,omitempty"`
	Tag           string  `protobuf:"bytes,6,opt,name=tag,proto3" json:"tag,omitempty"`
}

func (x *BoardTuple) Reset() {
	*x = BoardTuple{}
	if protoimpl.UnsafeEnabled {
		mi := &file_lib_assets_objective_objective_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BoardTuple) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BoardTuple) ProtoMessage() {}

func (x *BoardTuple) ProtoReflect() protoreflect.Message {
	mi := &file_lib_assets_objective_objective_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BoardTuple.ProtoReflect.Descriptor instead.
func (*BoardTuple) Descriptor() ([]byte, []int) {
	return file_lib_assets_objective_objective_proto_rawDescGZIP(), []int{4}
}

func (x *BoardTuple) GetAlgo() string {
	if x != nil {
		return x.Algo
	}
	return ""
}

func (x *BoardTuple) GetCreator() string {
	if x != nil {
		return x.Creator
	}
	return ""
}

func (x *BoardTuple) GetKey() string {
	if x != nil {
		return x.Key
	}
	return ""
}

func (x *BoardTuple) GetTraintupleKey() string {
	if x != nil {
		return x.TraintupleKey
	}
	return ""
}

func (x *BoardTuple) GetPerf() float32 {
	if x != nil {
		return x.Perf
	}
	return 0
}

func (x *BoardTuple) GetTag() string {
	if x != nil {
		return x.Tag
	}
	return ""
}

type Leaderboard struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Objective  *Objective    `protobuf:"bytes,1,opt,name=objective,proto3" json:"objective,omitempty"`
	TestTuples []*BoardTuple `protobuf:"bytes,2,rep,name=test_tuples,json=testTuples,proto3" json:"test_tuples,omitempty"`
}

func (x *Leaderboard) Reset() {
	*x = Leaderboard{}
	if protoimpl.UnsafeEnabled {
		mi := &file_lib_assets_objective_objective_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Leaderboard) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Leaderboard) ProtoMessage() {}

func (x *Leaderboard) ProtoReflect() protoreflect.Message {
	mi := &file_lib_assets_objective_objective_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Leaderboard.ProtoReflect.Descriptor instead.
func (*Leaderboard) Descriptor() ([]byte, []int) {
	return file_lib_assets_objective_objective_proto_rawDescGZIP(), []int{5}
}

func (x *Leaderboard) GetObjective() *Objective {
	if x != nil {
		return x.Objective
	}
	return nil
}

func (x *Leaderboard) GetTestTuples() []*BoardTuple {
	if x != nil {
		return x.TestTuples
	}
	return nil
}

var File_lib_assets_objective_objective_proto protoreflect.FileDescriptor

var file_lib_assets_objective_objective_proto_rawDesc = []byte{
	0x0a, 0x24, 0x6c, 0x69, 0x62, 0x2f, 0x61, 0x73, 0x73, 0x65, 0x74, 0x73, 0x2f, 0x6f, 0x62, 0x6a,
	0x65, 0x63, 0x74, 0x69, 0x76, 0x65, 0x2f, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x76, 0x65,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x09, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x76,
	0x65, 0x1a, 0x17, 0x6c, 0x69, 0x62, 0x2f, 0x61, 0x73, 0x73, 0x65, 0x74, 0x73, 0x2f, 0x63, 0x6f,
	0x6d, 0x6d, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x20, 0x6c, 0x69, 0x62, 0x2f,
	0x61, 0x73, 0x73, 0x65, 0x74, 0x73, 0x2f, 0x64, 0x61, 0x74, 0x61, 0x73, 0x65, 0x74, 0x2f, 0x64,
	0x61, 0x74, 0x61, 0x73, 0x65, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xa3, 0x03, 0x0a,
	0x09, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x76, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65,
	0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x12, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x12, 0x33, 0x0a, 0x0c, 0x74, 0x65, 0x73, 0x74, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x73, 0x65, 0x74,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x10, 0x2e, 0x64, 0x61, 0x74, 0x61, 0x73, 0x65, 0x74,
	0x2e, 0x44, 0x61, 0x74, 0x61, 0x73, 0x65, 0x74, 0x52, 0x0b, 0x74, 0x65, 0x73, 0x74, 0x44, 0x61,
	0x74, 0x61, 0x73, 0x65, 0x74, 0x12, 0x35, 0x0a, 0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70,
	0x74, 0x69, 0x6f, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x61, 0x73, 0x73,
	0x65, 0x74, 0x73, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x61, 0x62, 0x6c, 0x65, 0x52,
	0x0b, 0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x12, 0x21, 0x0a, 0x0c,
	0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x05, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x0b, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x4e, 0x61, 0x6d, 0x65, 0x12,
	0x2d, 0x0a, 0x07, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x13, 0x2e, 0x61, 0x73, 0x73, 0x65, 0x74, 0x73, 0x2e, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73,
	0x73, 0x61, 0x62, 0x6c, 0x65, 0x52, 0x07, 0x6d, 0x65, 0x74, 0x72, 0x69, 0x63, 0x73, 0x12, 0x35,
	0x0a, 0x0b, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x18, 0x07, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x13, 0x2e, 0x61, 0x73, 0x73, 0x65, 0x74, 0x73, 0x2e, 0x50, 0x65, 0x72,
	0x6d, 0x69, 0x73, 0x73, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x0b, 0x70, 0x65, 0x72, 0x6d, 0x69, 0x73,
	0x73, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x3e, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74,
	0x61, 0x18, 0x08, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x22, 0x2e, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74,
	0x69, 0x76, 0x65, 0x2e, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x76, 0x65, 0x2e, 0x4d, 0x65,
	0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x08, 0x6d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0x1a, 0x3b, 0x0a, 0x0d, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74,
	0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02,
	0x38, 0x01, 0x22, 0x15, 0x0a, 0x13, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x76, 0x65, 0x51,
	0x75, 0x65, 0x72, 0x79, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x22, 0x4e, 0x0a, 0x16, 0x4f, 0x62, 0x6a,
	0x65, 0x63, 0x74, 0x69, 0x76, 0x65, 0x51, 0x75, 0x65, 0x72, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x34, 0x0a, 0x0a, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x76, 0x65,
	0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74,
	0x69, 0x76, 0x65, 0x2e, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x76, 0x65, 0x52, 0x0a, 0x6f,
	0x62, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x76, 0x65, 0x73, 0x22, 0x6e, 0x0a, 0x15, 0x4c, 0x65, 0x61,
	0x64, 0x65, 0x72, 0x62, 0x6f, 0x61, 0x72, 0x64, 0x51, 0x75, 0x65, 0x72, 0x79, 0x50, 0x61, 0x72,
	0x61, 0x6d, 0x12, 0x23, 0x0a, 0x0d, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x76, 0x65, 0x5f,
	0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x6f, 0x62, 0x6a, 0x65, 0x63,
	0x74, 0x69, 0x76, 0x65, 0x4b, 0x65, 0x79, 0x12, 0x30, 0x0a, 0x0a, 0x73, 0x6f, 0x72, 0x74, 0x5f,
	0x6f, 0x72, 0x64, 0x65, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x11, 0x2e, 0x61, 0x73,
	0x73, 0x65, 0x74, 0x73, 0x2e, 0x53, 0x6f, 0x72, 0x74, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x52, 0x09,
	0x73, 0x6f, 0x72, 0x74, 0x4f, 0x72, 0x64, 0x65, 0x72, 0x22, 0x99, 0x01, 0x0a, 0x0a, 0x42, 0x6f,
	0x61, 0x72, 0x64, 0x54, 0x75, 0x70, 0x6c, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x61, 0x6c, 0x67, 0x6f,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x61, 0x6c, 0x67, 0x6f, 0x12, 0x18, 0x0a, 0x07,
	0x63, 0x72, 0x65, 0x61, 0x74, 0x6f, 0x72, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63,
	0x72, 0x65, 0x61, 0x74, 0x6f, 0x72, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x03, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x25, 0x0a, 0x0e, 0x74, 0x72, 0x61, 0x69,
	0x6e, 0x74, 0x75, 0x70, 0x6c, 0x65, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x0d, 0x74, 0x72, 0x61, 0x69, 0x6e, 0x74, 0x75, 0x70, 0x6c, 0x65, 0x4b, 0x65, 0x79, 0x12,
	0x12, 0x0a, 0x04, 0x70, 0x65, 0x72, 0x66, 0x18, 0x05, 0x20, 0x01, 0x28, 0x02, 0x52, 0x04, 0x70,
	0x65, 0x72, 0x66, 0x12, 0x10, 0x0a, 0x03, 0x74, 0x61, 0x67, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x03, 0x74, 0x61, 0x67, 0x22, 0x79, 0x0a, 0x0b, 0x4c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x62,
	0x6f, 0x61, 0x72, 0x64, 0x12, 0x32, 0x0a, 0x09, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x76,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x14, 0x2e, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74,
	0x69, 0x76, 0x65, 0x2e, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x76, 0x65, 0x52, 0x09, 0x6f,
	0x62, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x76, 0x65, 0x12, 0x36, 0x0a, 0x0b, 0x74, 0x65, 0x73, 0x74,
	0x5f, 0x74, 0x75, 0x70, 0x6c, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e,
	0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x76, 0x65, 0x2e, 0x42, 0x6f, 0x61, 0x72, 0x64, 0x54,
	0x75, 0x70, 0x6c, 0x65, 0x52, 0x0a, 0x74, 0x65, 0x73, 0x74, 0x54, 0x75, 0x70, 0x6c, 0x65, 0x73,
	0x32, 0xfb, 0x01, 0x0a, 0x10, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x76, 0x65, 0x53, 0x65,
	0x72, 0x76, 0x69, 0x63, 0x65, 0x12, 0x41, 0x0a, 0x11, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65,
	0x72, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x76, 0x65, 0x12, 0x14, 0x2e, 0x6f, 0x62, 0x6a,
	0x65, 0x63, 0x74, 0x69, 0x76, 0x65, 0x2e, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x76, 0x65,
	0x1a, 0x14, 0x2e, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x76, 0x65, 0x2e, 0x4f, 0x62, 0x6a,
	0x65, 0x63, 0x74, 0x69, 0x76, 0x65, 0x22, 0x00, 0x12, 0x56, 0x0a, 0x0f, 0x51, 0x75, 0x65, 0x72,
	0x79, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x76, 0x65, 0x73, 0x12, 0x1e, 0x2e, 0x6f, 0x62,
	0x6a, 0x65, 0x63, 0x74, 0x69, 0x76, 0x65, 0x2e, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x76,
	0x65, 0x51, 0x75, 0x65, 0x72, 0x79, 0x50, 0x61, 0x72, 0x61, 0x6d, 0x1a, 0x21, 0x2e, 0x6f, 0x62,
	0x6a, 0x65, 0x63, 0x74, 0x69, 0x76, 0x65, 0x2e, 0x4f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x76,
	0x65, 0x51, 0x75, 0x65, 0x72, 0x79, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00,
	0x12, 0x4c, 0x0a, 0x0e, 0x47, 0x65, 0x74, 0x4c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x62, 0x6f, 0x61,
	0x72, 0x64, 0x12, 0x20, 0x2e, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x76, 0x65, 0x2e, 0x4c,
	0x65, 0x61, 0x64, 0x65, 0x72, 0x62, 0x6f, 0x61, 0x72, 0x64, 0x51, 0x75, 0x65, 0x72, 0x79, 0x50,
	0x61, 0x72, 0x61, 0x6d, 0x1a, 0x16, 0x2e, 0x6f, 0x62, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x76, 0x65,
	0x2e, 0x4c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x62, 0x6f, 0x61, 0x72, 0x64, 0x22, 0x00, 0x42, 0x34,
	0x5a, 0x32, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x6f, 0x77, 0x6b,
	0x69, 0x6e, 0x2f, 0x6f, 0x72, 0x63, 0x68, 0x65, 0x73, 0x74, 0x72, 0x61, 0x74, 0x6f, 0x72, 0x2f,
	0x6c, 0x69, 0x62, 0x2f, 0x61, 0x73, 0x73, 0x65, 0x74, 0x73, 0x2f, 0x6f, 0x62, 0x6a, 0x65, 0x63,
	0x74, 0x69, 0x76, 0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_lib_assets_objective_objective_proto_rawDescOnce sync.Once
	file_lib_assets_objective_objective_proto_rawDescData = file_lib_assets_objective_objective_proto_rawDesc
)

func file_lib_assets_objective_objective_proto_rawDescGZIP() []byte {
	file_lib_assets_objective_objective_proto_rawDescOnce.Do(func() {
		file_lib_assets_objective_objective_proto_rawDescData = protoimpl.X.CompressGZIP(file_lib_assets_objective_objective_proto_rawDescData)
	})
	return file_lib_assets_objective_objective_proto_rawDescData
}

var file_lib_assets_objective_objective_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_lib_assets_objective_objective_proto_goTypes = []interface{}{
	(*Objective)(nil),              // 0: objective.Objective
	(*ObjectiveQueryParam)(nil),    // 1: objective.ObjectiveQueryParam
	(*ObjectiveQueryResponse)(nil), // 2: objective.ObjectiveQueryResponse
	(*LeaderboardQueryParam)(nil),  // 3: objective.LeaderboardQueryParam
	(*BoardTuple)(nil),             // 4: objective.BoardTuple
	(*Leaderboard)(nil),            // 5: objective.Leaderboard
	nil,                            // 6: objective.Objective.MetadataEntry
	(*dataset.Dataset)(nil),        // 7: dataset.Dataset
	(*assets.Addressable)(nil),     // 8: assets.Addressable
	(*assets.Permissions)(nil),     // 9: assets.Permissions
	(assets.SortOrder)(0),          // 10: assets.SortOrder
}
var file_lib_assets_objective_objective_proto_depIdxs = []int32{
	7,  // 0: objective.Objective.test_dataset:type_name -> dataset.Dataset
	8,  // 1: objective.Objective.description:type_name -> assets.Addressable
	8,  // 2: objective.Objective.metrics:type_name -> assets.Addressable
	9,  // 3: objective.Objective.permissions:type_name -> assets.Permissions
	6,  // 4: objective.Objective.metadata:type_name -> objective.Objective.MetadataEntry
	0,  // 5: objective.ObjectiveQueryResponse.objectives:type_name -> objective.Objective
	10, // 6: objective.LeaderboardQueryParam.sort_order:type_name -> assets.SortOrder
	0,  // 7: objective.Leaderboard.objective:type_name -> objective.Objective
	4,  // 8: objective.Leaderboard.test_tuples:type_name -> objective.BoardTuple
	0,  // 9: objective.ObjectiveService.RegisterObjective:input_type -> objective.Objective
	1,  // 10: objective.ObjectiveService.QueryObjectives:input_type -> objective.ObjectiveQueryParam
	3,  // 11: objective.ObjectiveService.GetLeaderboard:input_type -> objective.LeaderboardQueryParam
	0,  // 12: objective.ObjectiveService.RegisterObjective:output_type -> objective.Objective
	2,  // 13: objective.ObjectiveService.QueryObjectives:output_type -> objective.ObjectiveQueryResponse
	5,  // 14: objective.ObjectiveService.GetLeaderboard:output_type -> objective.Leaderboard
	12, // [12:15] is the sub-list for method output_type
	9,  // [9:12] is the sub-list for method input_type
	9,  // [9:9] is the sub-list for extension type_name
	9,  // [9:9] is the sub-list for extension extendee
	0,  // [0:9] is the sub-list for field type_name
}

func init() { file_lib_assets_objective_objective_proto_init() }
func file_lib_assets_objective_objective_proto_init() {
	if File_lib_assets_objective_objective_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_lib_assets_objective_objective_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Objective); i {
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
		file_lib_assets_objective_objective_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ObjectiveQueryParam); i {
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
		file_lib_assets_objective_objective_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ObjectiveQueryResponse); i {
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
		file_lib_assets_objective_objective_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LeaderboardQueryParam); i {
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
		file_lib_assets_objective_objective_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BoardTuple); i {
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
		file_lib_assets_objective_objective_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Leaderboard); i {
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
			RawDescriptor: file_lib_assets_objective_objective_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_lib_assets_objective_objective_proto_goTypes,
		DependencyIndexes: file_lib_assets_objective_objective_proto_depIdxs,
		MessageInfos:      file_lib_assets_objective_objective_proto_msgTypes,
	}.Build()
	File_lib_assets_objective_objective_proto = out.File
	file_lib_assets_objective_objective_proto_rawDesc = nil
	file_lib_assets_objective_objective_proto_goTypes = nil
	file_lib_assets_objective_objective_proto_depIdxs = nil
}
