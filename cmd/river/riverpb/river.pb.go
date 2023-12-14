// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.31.0
// 	protoc        (unknown)
// source: riverpb/river.proto

package riverpb

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	_ "google.golang.org/protobuf/types/known/anypb"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type RawData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Data []byte `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *RawData) Reset() {
	*x = RawData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_riverpb_river_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RawData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RawData) ProtoMessage() {}

func (x *RawData) ProtoReflect() protoreflect.Message {
	mi := &file_riverpb_river_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RawData.ProtoReflect.Descriptor instead.
func (*RawData) Descriptor() ([]byte, []int) {
	return file_riverpb_river_proto_rawDescGZIP(), []int{0}
}

func (x *RawData) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type Record struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key   []byte `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Value []byte `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
	Ttl   int64  `protobuf:"varint,3,opt,name=ttl,proto3" json:"ttl,omitempty"`
}

func (x *Record) Reset() {
	*x = Record{}
	if protoimpl.UnsafeEnabled {
		mi := &file_riverpb_river_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Record) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Record) ProtoMessage() {}

func (x *Record) ProtoReflect() protoreflect.Message {
	mi := &file_riverpb_river_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Record.ProtoReflect.Descriptor instead.
func (*Record) Descriptor() ([]byte, []int) {
	return file_riverpb_river_proto_rawDescGZIP(), []int{1}
}

func (x *Record) GetKey() []byte {
	if x != nil {
		return x.Key
	}
	return nil
}

func (x *Record) GetValue() []byte {
	if x != nil {
		return x.Value
	}
	return nil
}

func (x *Record) GetTtl() int64 {
	if x != nil {
		return x.Ttl
	}
	return 0
}

type ExpRecord struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Key []byte `protobuf:"bytes,1,opt,name=key,proto3" json:"key,omitempty"`
	Ttl int64  `protobuf:"varint,3,opt,name=ttl,proto3" json:"ttl,omitempty"`
}

func (x *ExpRecord) Reset() {
	*x = ExpRecord{}
	if protoimpl.UnsafeEnabled {
		mi := &file_riverpb_river_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ExpRecord) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ExpRecord) ProtoMessage() {}

func (x *ExpRecord) ProtoReflect() protoreflect.Message {
	mi := &file_riverpb_river_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ExpRecord.ProtoReflect.Descriptor instead.
func (*ExpRecord) Descriptor() ([]byte, []int) {
	return file_riverpb_river_proto_rawDescGZIP(), []int{2}
}

func (x *ExpRecord) GetKey() []byte {
	if x != nil {
		return x.Key
	}
	return nil
}

func (x *ExpRecord) GetTtl() int64 {
	if x != nil {
		return x.Ttl
	}
	return 0
}

type DataResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ok   bool   `protobuf:"varint,1,opt,name=ok,proto3" json:"ok,omitempty"`
	Data []byte `protobuf:"bytes,2,opt,name=data,proto3" json:"data,omitempty"`
}

func (x *DataResult) Reset() {
	*x = DataResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_riverpb_river_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DataResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DataResult) ProtoMessage() {}

func (x *DataResult) ProtoReflect() protoreflect.Message {
	mi := &file_riverpb_river_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DataResult.ProtoReflect.Descriptor instead.
func (*DataResult) Descriptor() ([]byte, []int) {
	return file_riverpb_river_proto_rawDescGZIP(), []int{3}
}

func (x *DataResult) GetOk() bool {
	if x != nil {
		return x.Ok
	}
	return false
}

func (x *DataResult) GetData() []byte {
	if x != nil {
		return x.Data
	}
	return nil
}

type TTLResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ok  bool  `protobuf:"varint,1,opt,name=ok,proto3" json:"ok,omitempty"`
	Ttl int64 `protobuf:"varint,2,opt,name=ttl,proto3" json:"ttl,omitempty"`
}

func (x *TTLResult) Reset() {
	*x = TTLResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_riverpb_river_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TTLResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TTLResult) ProtoMessage() {}

func (x *TTLResult) ProtoReflect() protoreflect.Message {
	mi := &file_riverpb_river_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TTLResult.ProtoReflect.Descriptor instead.
func (*TTLResult) Descriptor() ([]byte, []int) {
	return file_riverpb_river_proto_rawDescGZIP(), []int{4}
}

func (x *TTLResult) GetOk() bool {
	if x != nil {
		return x.Ok
	}
	return false
}

func (x *TTLResult) GetTtl() int64 {
	if x != nil {
		return x.Ttl
	}
	return 0
}

type InfoResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ok bool `protobuf:"varint,1,opt,name=ok,proto3" json:"ok,omitempty"`
}

func (x *InfoResult) Reset() {
	*x = InfoResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_riverpb_river_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *InfoResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*InfoResult) ProtoMessage() {}

func (x *InfoResult) ProtoReflect() protoreflect.Message {
	mi := &file_riverpb_river_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use InfoResult.ProtoReflect.Descriptor instead.
func (*InfoResult) Descriptor() ([]byte, []int) {
	return file_riverpb_river_proto_rawDescGZIP(), []int{5}
}

func (x *InfoResult) GetOk() bool {
	if x != nil {
		return x.Ok
	}
	return false
}

type Status struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Keys     int64 `protobuf:"varint,1,opt,name=keys,proto3" json:"keys,omitempty"`
	Records  int64 `protobuf:"varint,2,opt,name=records,proto3" json:"records,omitempty"`
	Datasize int64 `protobuf:"varint,3,opt,name=datasize,proto3" json:"datasize,omitempty"`
	Hintsize int64 `protobuf:"varint,4,opt,name=hintsize,proto3" json:"hintsize,omitempty"`
}

func (x *Status) Reset() {
	*x = Status{}
	if protoimpl.UnsafeEnabled {
		mi := &file_riverpb_river_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Status) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Status) ProtoMessage() {}

func (x *Status) ProtoReflect() protoreflect.Message {
	mi := &file_riverpb_river_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Status.ProtoReflect.Descriptor instead.
func (*Status) Descriptor() ([]byte, []int) {
	return file_riverpb_river_proto_rawDescGZIP(), []int{6}
}

func (x *Status) GetKeys() int64 {
	if x != nil {
		return x.Keys
	}
	return 0
}

func (x *Status) GetRecords() int64 {
	if x != nil {
		return x.Records
	}
	return 0
}

func (x *Status) GetDatasize() int64 {
	if x != nil {
		return x.Datasize
	}
	return 0
}

func (x *Status) GetHintsize() int64 {
	if x != nil {
		return x.Hintsize
	}
	return 0
}

type BatchPutOption struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Records   []*Record `protobuf:"bytes,1,rep,name=records,proto3" json:"records,omitempty"`
	BatchSize int64     `protobuf:"varint,2,opt,name=batchSize,proto3" json:"batchSize,omitempty"`
}

func (x *BatchPutOption) Reset() {
	*x = BatchPutOption{}
	if protoimpl.UnsafeEnabled {
		mi := &file_riverpb_river_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BatchPutOption) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BatchPutOption) ProtoMessage() {}

func (x *BatchPutOption) ProtoReflect() protoreflect.Message {
	mi := &file_riverpb_river_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BatchPutOption.ProtoReflect.Descriptor instead.
func (*BatchPutOption) Descriptor() ([]byte, []int) {
	return file_riverpb_river_proto_rawDescGZIP(), []int{7}
}

func (x *BatchPutOption) GetRecords() []*Record {
	if x != nil {
		return x.Records
	}
	return nil
}

func (x *BatchPutOption) GetBatchSize() int64 {
	if x != nil {
		return x.BatchSize
	}
	return 0
}

type BatchResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ok       bool  `protobuf:"varint,1,opt,name=ok,proto3" json:"ok,omitempty"`
	Effected int64 `protobuf:"varint,2,opt,name=effected,proto3" json:"effected,omitempty"`
}

func (x *BatchResult) Reset() {
	*x = BatchResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_riverpb_river_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BatchResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BatchResult) ProtoMessage() {}

func (x *BatchResult) ProtoReflect() protoreflect.Message {
	mi := &file_riverpb_river_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BatchResult.ProtoReflect.Descriptor instead.
func (*BatchResult) Descriptor() ([]byte, []int) {
	return file_riverpb_river_proto_rawDescGZIP(), []int{8}
}

func (x *BatchResult) GetOk() bool {
	if x != nil {
		return x.Ok
	}
	return false
}

func (x *BatchResult) GetEffected() int64 {
	if x != nil {
		return x.Effected
	}
	return 0
}

type BatchDelOption struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Keys      [][]byte `protobuf:"bytes,1,rep,name=keys,proto3" json:"keys,omitempty"`
	BatchSize int64    `protobuf:"varint,2,opt,name=batchSize,proto3" json:"batchSize,omitempty"`
}

func (x *BatchDelOption) Reset() {
	*x = BatchDelOption{}
	if protoimpl.UnsafeEnabled {
		mi := &file_riverpb_river_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BatchDelOption) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BatchDelOption) ProtoMessage() {}

func (x *BatchDelOption) ProtoReflect() protoreflect.Message {
	mi := &file_riverpb_river_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BatchDelOption.ProtoReflect.Descriptor instead.
func (*BatchDelOption) Descriptor() ([]byte, []int) {
	return file_riverpb_river_proto_rawDescGZIP(), []int{9}
}

func (x *BatchDelOption) GetKeys() [][]byte {
	if x != nil {
		return x.Keys
	}
	return nil
}

func (x *BatchDelOption) GetBatchSize() int64 {
	if x != nil {
		return x.BatchSize
	}
	return 0
}

type RangeOption struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	MinKey  []byte `protobuf:"bytes,1,opt,name=minKey,proto3" json:"minKey,omitempty"`
	MaxKey  []byte `protobuf:"bytes,2,opt,name=maxKey,proto3" json:"maxKey,omitempty"`
	Pattern []byte `protobuf:"bytes,3,opt,name=pattern,proto3" json:"pattern,omitempty"`
	Descend bool   `protobuf:"varint,4,opt,name=descend,proto3" json:"descend,omitempty"`
}

func (x *RangeOption) Reset() {
	*x = RangeOption{}
	if protoimpl.UnsafeEnabled {
		mi := &file_riverpb_river_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RangeOption) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RangeOption) ProtoMessage() {}

func (x *RangeOption) ProtoReflect() protoreflect.Message {
	mi := &file_riverpb_river_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RangeOption.ProtoReflect.Descriptor instead.
func (*RangeOption) Descriptor() ([]byte, []int) {
	return file_riverpb_river_proto_rawDescGZIP(), []int{10}
}

func (x *RangeOption) GetMinKey() []byte {
	if x != nil {
		return x.MinKey
	}
	return nil
}

func (x *RangeOption) GetMaxKey() []byte {
	if x != nil {
		return x.MaxKey
	}
	return nil
}

func (x *RangeOption) GetPattern() []byte {
	if x != nil {
		return x.Pattern
	}
	return nil
}

func (x *RangeOption) GetDescend() bool {
	if x != nil {
		return x.Descend
	}
	return false
}

type RangeResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Keys  [][]byte `protobuf:"bytes,1,rep,name=keys,proto3" json:"keys,omitempty"`
	Count int64    `protobuf:"varint,2,opt,name=count,proto3" json:"count,omitempty"`
}

func (x *RangeResult) Reset() {
	*x = RangeResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_riverpb_river_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RangeResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RangeResult) ProtoMessage() {}

func (x *RangeResult) ProtoReflect() protoreflect.Message {
	mi := &file_riverpb_river_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RangeResult.ProtoReflect.Descriptor instead.
func (*RangeResult) Descriptor() ([]byte, []int) {
	return file_riverpb_river_proto_rawDescGZIP(), []int{11}
}

func (x *RangeResult) GetKeys() [][]byte {
	if x != nil {
		return x.Keys
	}
	return nil
}

func (x *RangeResult) GetCount() int64 {
	if x != nil {
		return x.Count
	}
	return 0
}

var File_riverpb_river_proto protoreflect.FileDescriptor

var file_riverpb_river_proto_rawDesc = []byte{
	0x0a, 0x13, 0x72, 0x69, 0x76, 0x65, 0x72, 0x70, 0x62, 0x2f, 0x72, 0x69, 0x76, 0x65, 0x72, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x61, 0x6e, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x1b, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75,
	0x66, 0x2f, 0x65, 0x6d, 0x70, 0x74, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x1d, 0x0a,
	0x07, 0x52, 0x61, 0x77, 0x44, 0x61, 0x74, 0x61, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x22, 0x42, 0x0a, 0x06,
	0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x12, 0x10, 0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x0c, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x10,
	0x0a, 0x03, 0x74, 0x74, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x74, 0x74, 0x6c,
	0x22, 0x2f, 0x0a, 0x09, 0x45, 0x78, 0x70, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x10, 0x0a, 0x03, 0x74, 0x74, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x74, 0x74,
	0x6c, 0x22, 0x30, 0x0a, 0x0a, 0x44, 0x61, 0x74, 0x61, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12,
	0x0e, 0x0a, 0x02, 0x6f, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x02, 0x6f, 0x6b, 0x12,
	0x12, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x04, 0x64,
	0x61, 0x74, 0x61, 0x22, 0x2d, 0x0a, 0x09, 0x54, 0x54, 0x4c, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74,
	0x12, 0x0e, 0x0a, 0x02, 0x6f, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x02, 0x6f, 0x6b,
	0x12, 0x10, 0x0a, 0x03, 0x74, 0x74, 0x6c, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x03, 0x74,
	0x74, 0x6c, 0x22, 0x1c, 0x0a, 0x0a, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74,
	0x12, 0x0e, 0x0a, 0x02, 0x6f, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x02, 0x6f, 0x6b,
	0x22, 0x6e, 0x0a, 0x06, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x65,
	0x79, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x03, 0x52, 0x04, 0x6b, 0x65, 0x79, 0x73, 0x12, 0x18,
	0x0a, 0x07, 0x72, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52,
	0x07, 0x72, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x64, 0x61, 0x74, 0x61,
	0x73, 0x69, 0x7a, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x64, 0x61, 0x74, 0x61,
	0x73, 0x69, 0x7a, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x68, 0x69, 0x6e, 0x74, 0x73, 0x69, 0x7a, 0x65,
	0x18, 0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x68, 0x69, 0x6e, 0x74, 0x73, 0x69, 0x7a, 0x65,
	0x22, 0x51, 0x0a, 0x0e, 0x42, 0x61, 0x74, 0x63, 0x68, 0x50, 0x75, 0x74, 0x4f, 0x70, 0x74, 0x69,
	0x6f, 0x6e, 0x12, 0x21, 0x0a, 0x07, 0x72, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x73, 0x18, 0x01, 0x20,
	0x03, 0x28, 0x0b, 0x32, 0x07, 0x2e, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64, 0x52, 0x07, 0x72, 0x65,
	0x63, 0x6f, 0x72, 0x64, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x62, 0x61, 0x74, 0x63, 0x68, 0x53, 0x69,
	0x7a, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x62, 0x61, 0x74, 0x63, 0x68, 0x53,
	0x69, 0x7a, 0x65, 0x22, 0x39, 0x0a, 0x0b, 0x42, 0x61, 0x74, 0x63, 0x68, 0x52, 0x65, 0x73, 0x75,
	0x6c, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x6f, 0x6b, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x02,
	0x6f, 0x6b, 0x12, 0x1a, 0x0a, 0x08, 0x65, 0x66, 0x66, 0x65, 0x63, 0x74, 0x65, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x65, 0x66, 0x66, 0x65, 0x63, 0x74, 0x65, 0x64, 0x22, 0x42,
	0x0a, 0x0e, 0x42, 0x61, 0x74, 0x63, 0x68, 0x44, 0x65, 0x6c, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x12, 0x0a, 0x04, 0x6b, 0x65, 0x79, 0x73, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0c, 0x52, 0x04,
	0x6b, 0x65, 0x79, 0x73, 0x12, 0x1c, 0x0a, 0x09, 0x62, 0x61, 0x74, 0x63, 0x68, 0x53, 0x69, 0x7a,
	0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x09, 0x62, 0x61, 0x74, 0x63, 0x68, 0x53, 0x69,
	0x7a, 0x65, 0x22, 0x71, 0x0a, 0x0b, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f,
	0x6e, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x69, 0x6e, 0x4b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0c, 0x52, 0x06, 0x6d, 0x69, 0x6e, 0x4b, 0x65, 0x79, 0x12, 0x16, 0x0a, 0x06, 0x6d, 0x61, 0x78,
	0x4b, 0x65, 0x79, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x06, 0x6d, 0x61, 0x78, 0x4b, 0x65,
	0x79, 0x12, 0x18, 0x0a, 0x07, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x07, 0x70, 0x61, 0x74, 0x74, 0x65, 0x72, 0x6e, 0x12, 0x18, 0x0a, 0x07, 0x64,
	0x65, 0x73, 0x63, 0x65, 0x6e, 0x64, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x64, 0x65,
	0x73, 0x63, 0x65, 0x6e, 0x64, 0x22, 0x37, 0x0a, 0x0b, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x52, 0x65,
	0x73, 0x75, 0x6c, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6b, 0x65, 0x79, 0x73, 0x18, 0x01, 0x20, 0x03,
	0x28, 0x0c, 0x52, 0x04, 0x6b, 0x65, 0x79, 0x73, 0x12, 0x14, 0x0a, 0x05, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x32, 0xc5,
	0x02, 0x0a, 0x05, 0x52, 0x69, 0x76, 0x65, 0x72, 0x12, 0x1c, 0x0a, 0x03, 0x47, 0x65, 0x74, 0x12,
	0x08, 0x2e, 0x52, 0x61, 0x77, 0x44, 0x61, 0x74, 0x61, 0x1a, 0x0b, 0x2e, 0x44, 0x61, 0x74, 0x61,
	0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x1b, 0x0a, 0x03, 0x54, 0x54, 0x4c, 0x12, 0x08, 0x2e,
	0x52, 0x61, 0x77, 0x44, 0x61, 0x74, 0x61, 0x1a, 0x0a, 0x2e, 0x54, 0x54, 0x4c, 0x52, 0x65, 0x73,
	0x75, 0x6c, 0x74, 0x12, 0x1b, 0x0a, 0x03, 0x50, 0x75, 0x74, 0x12, 0x07, 0x2e, 0x52, 0x65, 0x63,
	0x6f, 0x72, 0x64, 0x1a, 0x0b, 0x2e, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74,
	0x12, 0x2b, 0x0a, 0x0a, 0x50, 0x75, 0x74, 0x49, 0x6e, 0x42, 0x61, 0x74, 0x63, 0x68, 0x12, 0x0f,
	0x2e, 0x42, 0x61, 0x74, 0x63, 0x68, 0x50, 0x75, 0x74, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x1a,
	0x0c, 0x2e, 0x42, 0x61, 0x74, 0x63, 0x68, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x1e, 0x0a,
	0x03, 0x45, 0x78, 0x70, 0x12, 0x0a, 0x2e, 0x45, 0x78, 0x70, 0x52, 0x65, 0x63, 0x6f, 0x72, 0x64,
	0x1a, 0x0b, 0x2e, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x1c, 0x0a,
	0x03, 0x44, 0x65, 0x6c, 0x12, 0x08, 0x2e, 0x52, 0x61, 0x77, 0x44, 0x61, 0x74, 0x61, 0x1a, 0x0b,
	0x2e, 0x49, 0x6e, 0x66, 0x6f, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x2b, 0x0a, 0x0a, 0x44,
	0x65, 0x6c, 0x49, 0x6e, 0x42, 0x61, 0x74, 0x63, 0x68, 0x12, 0x0f, 0x2e, 0x42, 0x61, 0x74, 0x63,
	0x68, 0x44, 0x65, 0x6c, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x1a, 0x0c, 0x2e, 0x42, 0x61, 0x74,
	0x63, 0x68, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x27, 0x0a, 0x04, 0x53, 0x74, 0x61, 0x74,
	0x12, 0x16, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62,
	0x75, 0x66, 0x2e, 0x45, 0x6d, 0x70, 0x74, 0x79, 0x1a, 0x07, 0x2e, 0x53, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x12, 0x23, 0x0a, 0x05, 0x52, 0x61, 0x6e, 0x67, 0x65, 0x12, 0x0c, 0x2e, 0x52, 0x61, 0x6e,
	0x67, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x1a, 0x0c, 0x2e, 0x52, 0x61, 0x6e, 0x67, 0x65,
	0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x42, 0x0b, 0x5a, 0x09, 0x2e, 0x3b, 0x72, 0x69, 0x76, 0x65,
	0x72, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_riverpb_river_proto_rawDescOnce sync.Once
	file_riverpb_river_proto_rawDescData = file_riverpb_river_proto_rawDesc
)

func file_riverpb_river_proto_rawDescGZIP() []byte {
	file_riverpb_river_proto_rawDescOnce.Do(func() {
		file_riverpb_river_proto_rawDescData = protoimpl.X.CompressGZIP(file_riverpb_river_proto_rawDescData)
	})
	return file_riverpb_river_proto_rawDescData
}

var file_riverpb_river_proto_msgTypes = make([]protoimpl.MessageInfo, 12)
var file_riverpb_river_proto_goTypes = []interface{}{
	(*RawData)(nil),        // 0: RawData
	(*Record)(nil),         // 1: Record
	(*ExpRecord)(nil),      // 2: ExpRecord
	(*DataResult)(nil),     // 3: DataResult
	(*TTLResult)(nil),      // 4: TTLResult
	(*InfoResult)(nil),     // 5: InfoResult
	(*Status)(nil),         // 6: Status
	(*BatchPutOption)(nil), // 7: BatchPutOption
	(*BatchResult)(nil),    // 8: BatchResult
	(*BatchDelOption)(nil), // 9: BatchDelOption
	(*RangeOption)(nil),    // 10: RangeOption
	(*RangeResult)(nil),    // 11: RangeResult
	(*emptypb.Empty)(nil),  // 12: google.protobuf.Empty
}
var file_riverpb_river_proto_depIdxs = []int32{
	1,  // 0: BatchPutOption.records:type_name -> Record
	0,  // 1: River.Get:input_type -> RawData
	0,  // 2: River.TTL:input_type -> RawData
	1,  // 3: River.Put:input_type -> Record
	7,  // 4: River.PutInBatch:input_type -> BatchPutOption
	2,  // 5: River.Exp:input_type -> ExpRecord
	0,  // 6: River.Del:input_type -> RawData
	9,  // 7: River.DelInBatch:input_type -> BatchDelOption
	12, // 8: River.Stat:input_type -> google.protobuf.Empty
	10, // 9: River.Range:input_type -> RangeOption
	3,  // 10: River.Get:output_type -> DataResult
	4,  // 11: River.TTL:output_type -> TTLResult
	5,  // 12: River.Put:output_type -> InfoResult
	8,  // 13: River.PutInBatch:output_type -> BatchResult
	5,  // 14: River.Exp:output_type -> InfoResult
	5,  // 15: River.Del:output_type -> InfoResult
	8,  // 16: River.DelInBatch:output_type -> BatchResult
	6,  // 17: River.Stat:output_type -> Status
	11, // 18: River.Range:output_type -> RangeResult
	10, // [10:19] is the sub-list for method output_type
	1,  // [1:10] is the sub-list for method input_type
	1,  // [1:1] is the sub-list for extension type_name
	1,  // [1:1] is the sub-list for extension extendee
	0,  // [0:1] is the sub-list for field type_name
}

func init() { file_riverpb_river_proto_init() }
func file_riverpb_river_proto_init() {
	if File_riverpb_river_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_riverpb_river_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RawData); i {
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
		file_riverpb_river_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Record); i {
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
		file_riverpb_river_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ExpRecord); i {
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
		file_riverpb_river_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DataResult); i {
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
		file_riverpb_river_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TTLResult); i {
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
		file_riverpb_river_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*InfoResult); i {
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
		file_riverpb_river_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Status); i {
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
		file_riverpb_river_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BatchPutOption); i {
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
		file_riverpb_river_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BatchResult); i {
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
		file_riverpb_river_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BatchDelOption); i {
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
		file_riverpb_river_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RangeOption); i {
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
		file_riverpb_river_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RangeResult); i {
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
			RawDescriptor: file_riverpb_river_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   12,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_riverpb_river_proto_goTypes,
		DependencyIndexes: file_riverpb_river_proto_depIdxs,
		MessageInfos:      file_riverpb_river_proto_msgTypes,
	}.Build()
	File_riverpb_river_proto = out.File
	file_riverpb_river_proto_rawDesc = nil
	file_riverpb_river_proto_goTypes = nil
	file_riverpb_river_proto_depIdxs = nil
}
