// to generate proto use:  `protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative pkg/gRPC/event.proto`

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.12
// source: pkg/gRPC/event.proto

package gRPC

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type OnCreate struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Instance *Instance `protobuf:"bytes,1,opt,name=instance,proto3" json:"instance,omitempty"`
}

func (x *OnCreate) Reset() {
	*x = OnCreate{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_gRPC_event_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OnCreate) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OnCreate) ProtoMessage() {}

func (x *OnCreate) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_gRPC_event_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OnCreate.ProtoReflect.Descriptor instead.
func (*OnCreate) Descriptor() ([]byte, []int) {
	return file_pkg_gRPC_event_proto_rawDescGZIP(), []int{0}
}

func (x *OnCreate) GetInstance() *Instance {
	if x != nil {
		return x.Instance
	}
	return nil
}

type OnShutDown struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Instance *Instance `protobuf:"bytes,1,opt,name=instance,proto3" json:"instance,omitempty"`
}

func (x *OnShutDown) Reset() {
	*x = OnShutDown{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_gRPC_event_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OnShutDown) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OnShutDown) ProtoMessage() {}

func (x *OnShutDown) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_gRPC_event_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OnShutDown.ProtoReflect.Descriptor instead.
func (*OnShutDown) Descriptor() ([]byte, []int) {
	return file_pkg_gRPC_event_proto_rawDescGZIP(), []int{1}
}

func (x *OnShutDown) GetInstance() *Instance {
	if x != nil {
		return x.Instance
	}
	return nil
}

type OnHealthCheck struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Instance *Instance `protobuf:"bytes,1,opt,name=instance,proto3" json:"instance,omitempty"`
}

func (x *OnHealthCheck) Reset() {
	*x = OnHealthCheck{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_gRPC_event_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *OnHealthCheck) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*OnHealthCheck) ProtoMessage() {}

func (x *OnHealthCheck) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_gRPC_event_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use OnHealthCheck.ProtoReflect.Descriptor instead.
func (*OnHealthCheck) Descriptor() ([]byte, []int) {
	return file_pkg_gRPC_event_proto_rawDescGZIP(), []int{2}
}

func (x *OnHealthCheck) GetInstance() *Instance {
	if x != nil {
		return x.Instance
	}
	return nil
}

type Response struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success bool `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
}

func (x *Response) Reset() {
	*x = Response{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_gRPC_event_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Response) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Response) ProtoMessage() {}

func (x *Response) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_gRPC_event_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Response.ProtoReflect.Descriptor instead.
func (*Response) Descriptor() ([]byte, []int) {
	return file_pkg_gRPC_event_proto_rawDescGZIP(), []int{3}
}

func (x *Response) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

type Instance struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Role      int32                  `protobuf:"varint,1,opt,name=role,proto3" json:"role,omitempty"`
	Id        string                 `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	Ip        string                 `protobuf:"bytes,3,opt,name=ip,proto3" json:"ip,omitempty"`
	Cluster   string                 `protobuf:"bytes,4,opt,name=cluster,proto3" json:"cluster,omitempty"`
	CreatedAt *timestamppb.Timestamp `protobuf:"bytes,5,opt,name=createdAt,proto3" json:"createdAt,omitempty"`
}

func (x *Instance) Reset() {
	*x = Instance{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_gRPC_event_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Instance) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Instance) ProtoMessage() {}

func (x *Instance) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_gRPC_event_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Instance.ProtoReflect.Descriptor instead.
func (*Instance) Descriptor() ([]byte, []int) {
	return file_pkg_gRPC_event_proto_rawDescGZIP(), []int{4}
}

func (x *Instance) GetRole() int32 {
	if x != nil {
		return x.Role
	}
	return 0
}

func (x *Instance) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *Instance) GetIp() string {
	if x != nil {
		return x.Ip
	}
	return ""
}

func (x *Instance) GetCluster() string {
	if x != nil {
		return x.Cluster
	}
	return ""
}

func (x *Instance) GetCreatedAt() *timestamppb.Timestamp {
	if x != nil {
		return x.CreatedAt
	}
	return nil
}

type RoleData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Role int32 `protobuf:"varint,1,opt,name=role,proto3" json:"role,omitempty"`
}

func (x *RoleData) Reset() {
	*x = RoleData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_gRPC_event_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RoleData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RoleData) ProtoMessage() {}

func (x *RoleData) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_gRPC_event_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RoleData.ProtoReflect.Descriptor instead.
func (*RoleData) Descriptor() ([]byte, []int) {
	return file_pkg_gRPC_event_proto_rawDescGZIP(), []int{5}
}

func (x *RoleData) GetRole() int32 {
	if x != nil {
		return x.Role
	}
	return 0
}

type Instances struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Items []*Instance `protobuf:"bytes,1,rep,name=items,proto3" json:"items,omitempty"`
}

func (x *Instances) Reset() {
	*x = Instances{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pkg_gRPC_event_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Instances) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Instances) ProtoMessage() {}

func (x *Instances) ProtoReflect() protoreflect.Message {
	mi := &file_pkg_gRPC_event_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Instances.ProtoReflect.Descriptor instead.
func (*Instances) Descriptor() ([]byte, []int) {
	return file_pkg_gRPC_event_proto_rawDescGZIP(), []int{6}
}

func (x *Instances) GetItems() []*Instance {
	if x != nil {
		return x.Items
	}
	return nil
}

var File_pkg_gRPC_event_proto protoreflect.FileDescriptor

var file_pkg_gRPC_event_proto_rawDesc = []byte{
	0x0a, 0x14, 0x70, 0x6b, 0x67, 0x2f, 0x67, 0x52, 0x50, 0x43, 0x2f, 0x65, 0x76, 0x65, 0x6e, 0x74,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x04, 0x67, 0x52, 0x50, 0x43, 0x1a, 0x1f, 0x67, 0x6f,
	0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x74, 0x69,
	0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x36, 0x0a,
	0x08, 0x4f, 0x6e, 0x43, 0x72, 0x65, 0x61, 0x74, 0x65, 0x12, 0x2a, 0x0a, 0x08, 0x69, 0x6e, 0x73,
	0x74, 0x61, 0x6e, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x67, 0x52,
	0x50, 0x43, 0x2e, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x52, 0x08, 0x69, 0x6e, 0x73,
	0x74, 0x61, 0x6e, 0x63, 0x65, 0x22, 0x38, 0x0a, 0x0a, 0x4f, 0x6e, 0x53, 0x68, 0x75, 0x74, 0x44,
	0x6f, 0x77, 0x6e, 0x12, 0x2a, 0x0a, 0x08, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x67, 0x52, 0x50, 0x43, 0x2e, 0x49, 0x6e, 0x73,
	0x74, 0x61, 0x6e, 0x63, 0x65, 0x52, 0x08, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x22,
	0x3b, 0x0a, 0x0d, 0x4f, 0x6e, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x43, 0x68, 0x65, 0x63, 0x6b,
	0x12, 0x2a, 0x0a, 0x08, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x67, 0x52, 0x50, 0x43, 0x2e, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e,
	0x63, 0x65, 0x52, 0x08, 0x69, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x22, 0x24, 0x0a, 0x08,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63,
	0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65,
	0x73, 0x73, 0x22, 0x92, 0x01, 0x0a, 0x08, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65, 0x12,
	0x12, 0x0a, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x72,
	0x6f, 0x6c, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x02, 0x69, 0x64, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x70, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x02, 0x69, 0x70, 0x12, 0x18, 0x0a, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x12, 0x38, 0x0a,
	0x09, 0x63, 0x72, 0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0b,
	0x32, 0x1a, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x54, 0x69, 0x6d, 0x65, 0x73, 0x74, 0x61, 0x6d, 0x70, 0x52, 0x09, 0x63, 0x72,
	0x65, 0x61, 0x74, 0x65, 0x64, 0x41, 0x74, 0x22, 0x1e, 0x0a, 0x08, 0x52, 0x6f, 0x6c, 0x65, 0x44,
	0x61, 0x74, 0x61, 0x12, 0x12, 0x0a, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x05, 0x52, 0x04, 0x72, 0x6f, 0x6c, 0x65, 0x22, 0x31, 0x0a, 0x09, 0x49, 0x6e, 0x73, 0x74, 0x61,
	0x6e, 0x63, 0x65, 0x73, 0x12, 0x24, 0x0a, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x67, 0x52, 0x50, 0x43, 0x2e, 0x49, 0x6e, 0x73, 0x74, 0x61,
	0x6e, 0x63, 0x65, 0x52, 0x05, 0x69, 0x74, 0x65, 0x6d, 0x73, 0x32, 0xd7, 0x01, 0x0a, 0x05, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x12, 0x2f, 0x0a, 0x0b, 0x41, 0x66, 0x74, 0x65, 0x72, 0x43, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x12, 0x0e, 0x2e, 0x67, 0x52, 0x50, 0x43, 0x2e, 0x4f, 0x6e, 0x43, 0x72, 0x65,
	0x61, 0x74, 0x65, 0x1a, 0x0e, 0x2e, 0x67, 0x52, 0x50, 0x43, 0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x34, 0x0a, 0x0e, 0x42, 0x65, 0x66, 0x6f, 0x72, 0x65, 0x53,
	0x68, 0x75, 0x74, 0x44, 0x6f, 0x77, 0x6e, 0x12, 0x10, 0x2e, 0x67, 0x52, 0x50, 0x43, 0x2e, 0x4f,
	0x6e, 0x53, 0x68, 0x75, 0x74, 0x44, 0x6f, 0x77, 0x6e, 0x1a, 0x0e, 0x2e, 0x67, 0x52, 0x50, 0x43,
	0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x34, 0x0a, 0x0b, 0x48,
	0x65, 0x61, 0x6c, 0x74, 0x68, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x12, 0x13, 0x2e, 0x67, 0x52, 0x50,
	0x43, 0x2e, 0x4f, 0x6e, 0x48, 0x65, 0x61, 0x6c, 0x74, 0x68, 0x43, 0x68, 0x65, 0x63, 0x6b, 0x1a,
	0x0e, 0x2e, 0x67, 0x52, 0x50, 0x43, 0x2e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22,
	0x00, 0x12, 0x31, 0x0a, 0x0c, 0x47, 0x65, 0x74, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63, 0x65,
	0x73, 0x12, 0x0e, 0x2e, 0x67, 0x52, 0x50, 0x43, 0x2e, 0x52, 0x6f, 0x6c, 0x65, 0x44, 0x61, 0x74,
	0x61, 0x1a, 0x0f, 0x2e, 0x67, 0x52, 0x50, 0x43, 0x2e, 0x49, 0x6e, 0x73, 0x74, 0x61, 0x6e, 0x63,
	0x65, 0x73, 0x22, 0x00, 0x42, 0x28, 0x5a, 0x26, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x6e, 0x65, 0x6f, 0x6e, 0x6c, 0x61, 0x62, 0x73, 0x6f, 0x72, 0x67, 0x2f, 0x6e,
	0x65, 0x6f, 0x6e, 0x2d, 0x70, 0x72, 0x6f, 0x78, 0x79, 0x3b, 0x67, 0x52, 0x50, 0x43, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_pkg_gRPC_event_proto_rawDescOnce sync.Once
	file_pkg_gRPC_event_proto_rawDescData = file_pkg_gRPC_event_proto_rawDesc
)

func file_pkg_gRPC_event_proto_rawDescGZIP() []byte {
	file_pkg_gRPC_event_proto_rawDescOnce.Do(func() {
		file_pkg_gRPC_event_proto_rawDescData = protoimpl.X.CompressGZIP(file_pkg_gRPC_event_proto_rawDescData)
	})
	return file_pkg_gRPC_event_proto_rawDescData
}

var file_pkg_gRPC_event_proto_msgTypes = make([]protoimpl.MessageInfo, 7)
var file_pkg_gRPC_event_proto_goTypes = []interface{}{
	(*OnCreate)(nil),              // 0: gRPC.OnCreate
	(*OnShutDown)(nil),            // 1: gRPC.OnShutDown
	(*OnHealthCheck)(nil),         // 2: gRPC.OnHealthCheck
	(*Response)(nil),              // 3: gRPC.Response
	(*Instance)(nil),              // 4: gRPC.Instance
	(*RoleData)(nil),              // 5: gRPC.RoleData
	(*Instances)(nil),             // 6: gRPC.Instances
	(*timestamppb.Timestamp)(nil), // 7: google.protobuf.Timestamp
}
var file_pkg_gRPC_event_proto_depIdxs = []int32{
	4, // 0: gRPC.OnCreate.instance:type_name -> gRPC.Instance
	4, // 1: gRPC.OnShutDown.instance:type_name -> gRPC.Instance
	4, // 2: gRPC.OnHealthCheck.instance:type_name -> gRPC.Instance
	7, // 3: gRPC.Instance.createdAt:type_name -> google.protobuf.Timestamp
	4, // 4: gRPC.Instances.items:type_name -> gRPC.Instance
	0, // 5: gRPC.Event.AfterCreate:input_type -> gRPC.OnCreate
	1, // 6: gRPC.Event.BeforeShutDown:input_type -> gRPC.OnShutDown
	2, // 7: gRPC.Event.HealthCheck:input_type -> gRPC.OnHealthCheck
	5, // 8: gRPC.Event.GetInstances:input_type -> gRPC.RoleData
	3, // 9: gRPC.Event.AfterCreate:output_type -> gRPC.Response
	3, // 10: gRPC.Event.BeforeShutDown:output_type -> gRPC.Response
	3, // 11: gRPC.Event.HealthCheck:output_type -> gRPC.Response
	6, // 12: gRPC.Event.GetInstances:output_type -> gRPC.Instances
	9, // [9:13] is the sub-list for method output_type
	5, // [5:9] is the sub-list for method input_type
	5, // [5:5] is the sub-list for extension type_name
	5, // [5:5] is the sub-list for extension extendee
	0, // [0:5] is the sub-list for field type_name
}

func init() { file_pkg_gRPC_event_proto_init() }
func file_pkg_gRPC_event_proto_init() {
	if File_pkg_gRPC_event_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pkg_gRPC_event_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OnCreate); i {
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
		file_pkg_gRPC_event_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OnShutDown); i {
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
		file_pkg_gRPC_event_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*OnHealthCheck); i {
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
		file_pkg_gRPC_event_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Response); i {
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
		file_pkg_gRPC_event_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Instance); i {
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
		file_pkg_gRPC_event_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RoleData); i {
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
		file_pkg_gRPC_event_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Instances); i {
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
			RawDescriptor: file_pkg_gRPC_event_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   7,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pkg_gRPC_event_proto_goTypes,
		DependencyIndexes: file_pkg_gRPC_event_proto_depIdxs,
		MessageInfos:      file_pkg_gRPC_event_proto_msgTypes,
	}.Build()
	File_pkg_gRPC_event_proto = out.File
	file_pkg_gRPC_event_proto_rawDesc = nil
	file_pkg_gRPC_event_proto_goTypes = nil
	file_pkg_gRPC_event_proto_depIdxs = nil
}
