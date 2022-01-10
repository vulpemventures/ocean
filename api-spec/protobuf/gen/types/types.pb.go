// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        (unknown)
// source: types/types.proto

package types

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

type TxEventType int32

const (
	// Tx broadcasted.
	TxEventType_BROADCASTED TxEventType = 0
	// Tx unconfirmed.
	TxEventType_UNCONFIRMED TxEventType = 2
	// Tx confirmed.
	TxEventType_CONFIRMED TxEventType = 3
)

// Enum value maps for TxEventType.
var (
	TxEventType_name = map[int32]string{
		0: "BROADCASTED",
		2: "UNCONFIRMED",
		3: "CONFIRMED",
	}
	TxEventType_value = map[string]int32{
		"BROADCASTED": 0,
		"UNCONFIRMED": 2,
		"CONFIRMED":   3,
	}
)

func (x TxEventType) Enum() *TxEventType {
	p := new(TxEventType)
	*p = x
	return p
}

func (x TxEventType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (TxEventType) Descriptor() protoreflect.EnumDescriptor {
	return file_types_types_proto_enumTypes[0].Descriptor()
}

func (TxEventType) Type() protoreflect.EnumType {
	return &file_types_types_proto_enumTypes[0]
}

func (x TxEventType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use TxEventType.Descriptor instead.
func (TxEventType) EnumDescriptor() ([]byte, []int) {
	return file_types_types_proto_rawDescGZIP(), []int{0}
}

type Unspent_UnspentStatus int32

const (
	Unspent_CONFIRMED   Unspent_UnspentStatus = 0
	Unspent_UNCONFIRMED Unspent_UnspentStatus = 1
	Unspent_LOCKED      Unspent_UnspentStatus = 2
	Unspent_UNLOCKED    Unspent_UnspentStatus = 3
)

// Enum value maps for Unspent_UnspentStatus.
var (
	Unspent_UnspentStatus_name = map[int32]string{
		0: "CONFIRMED",
		1: "UNCONFIRMED",
		2: "LOCKED",
		3: "UNLOCKED",
	}
	Unspent_UnspentStatus_value = map[string]int32{
		"CONFIRMED":   0,
		"UNCONFIRMED": 1,
		"LOCKED":      2,
		"UNLOCKED":    3,
	}
)

func (x Unspent_UnspentStatus) Enum() *Unspent_UnspentStatus {
	p := new(Unspent_UnspentStatus)
	*p = x
	return p
}

func (x Unspent_UnspentStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Unspent_UnspentStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_types_types_proto_enumTypes[1].Descriptor()
}

func (Unspent_UnspentStatus) Type() protoreflect.EnumType {
	return &file_types_types_proto_enumTypes[1]
}

func (x Unspent_UnspentStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Unspent_UnspentStatus.Descriptor instead.
func (Unspent_UnspentStatus) EnumDescriptor() ([]byte, []int) {
	return file_types_types_proto_rawDescGZIP(), []int{2, 0}
}

type Template_Format int32

const (
	Template_DESCRIPTOR Template_Format = 0
	Template_MINISCRIPT Template_Format = 1
	Template_IONIO      Template_Format = 2
	Template_RAW        Template_Format = 3
)

// Enum value maps for Template_Format.
var (
	Template_Format_name = map[int32]string{
		0: "DESCRIPTOR",
		1: "MINISCRIPT",
		2: "IONIO",
		3: "RAW",
	}
	Template_Format_value = map[string]int32{
		"DESCRIPTOR": 0,
		"MINISCRIPT": 1,
		"IONIO":      2,
		"RAW":        3,
	}
)

func (x Template_Format) Enum() *Template_Format {
	p := new(Template_Format)
	*p = x
	return p
}

func (x Template_Format) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Template_Format) Descriptor() protoreflect.EnumDescriptor {
	return file_types_types_proto_enumTypes[2].Descriptor()
}

func (Template_Format) Type() protoreflect.EnumType {
	return &file_types_types_proto_enumTypes[2]
}

func (x Template_Format) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Template_Format.Descriptor instead.
func (Template_Format) EnumDescriptor() ([]byte, []int) {
	return file_types_types_proto_rawDescGZIP(), []int{3, 0}
}

type AccountKey struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Id of the account.
	Id int64 `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
	// Name of the account to be updated.
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *AccountKey) Reset() {
	*x = AccountKey{}
	if protoimpl.UnsafeEnabled {
		mi := &file_types_types_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AccountKey) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AccountKey) ProtoMessage() {}

func (x *AccountKey) ProtoReflect() protoreflect.Message {
	mi := &file_types_types_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AccountKey.ProtoReflect.Descriptor instead.
func (*AccountKey) Descriptor() ([]byte, []int) {
	return file_types_types_proto_rawDescGZIP(), []int{0}
}

func (x *AccountKey) GetId() int64 {
	if x != nil {
		return x.Id
	}
	return 0
}

func (x *AccountKey) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type Unspents struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// account key
	AccountKey *AccountKey `protobuf:"bytes,1,opt,name=account_key,json=accountKey,proto3" json:"account_key,omitempty"`
	// list of unspents
	Unspents []*Unspent `protobuf:"bytes,2,rep,name=unspents,proto3" json:"unspents,omitempty"`
}

func (x *Unspents) Reset() {
	*x = Unspents{}
	if protoimpl.UnsafeEnabled {
		mi := &file_types_types_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Unspents) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Unspents) ProtoMessage() {}

func (x *Unspents) ProtoReflect() protoreflect.Message {
	mi := &file_types_types_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Unspents.ProtoReflect.Descriptor instead.
func (*Unspents) Descriptor() ([]byte, []int) {
	return file_types_types_proto_rawDescGZIP(), []int{1}
}

func (x *Unspents) GetAccountKey() *AccountKey {
	if x != nil {
		return x.AccountKey
	}
	return nil
}

func (x *Unspents) GetUnspents() []*Unspent {
	if x != nil {
		return x.Unspents
	}
	return nil
}

type Unspent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Txid of the unspent
	Txid string `protobuf:"bytes,1,opt,name=txid,proto3" json:"txid,omitempty"`
	// Output index
	Index int64 `protobuf:"varint,2,opt,name=index,proto3" json:"index,omitempty"`
	// Asset
	Asset []byte `protobuf:"bytes,3,opt,name=asset,proto3" json:"asset,omitempty"`
	// Asset commitment, empty if asset is not confidential
	AssetCommitment []byte `protobuf:"bytes,4,opt,name=asset_commitment,json=assetCommitment,proto3" json:"asset_commitment,omitempty"`
	// Value
	Value []byte `protobuf:"bytes,5,opt,name=value,proto3" json:"value,omitempty"`
	// Value commitment, empty if value is not confidential
	ValueCommitment []byte `protobuf:"bytes,6,opt,name=value_commitment,json=valueCommitment,proto3" json:"value_commitment,omitempty"`
	// Script
	Script []byte `protobuf:"bytes,7,opt,name=script,proto3" json:"script,omitempty"`
	// Nonce
	Nonce []byte `protobuf:"bytes,8,opt,name=nonce,proto3" json:"nonce,omitempty"`
	// Range proof
	RangeProof []byte `protobuf:"bytes,9,opt,name=rangeProof,proto3" json:"rangeProof,omitempty"`
	// Surjection proof
	SurjectionProof []byte `protobuf:"bytes,10,opt,name=surjectionProof,proto3" json:"surjectionProof,omitempty"`
	// Unspent status.
	Status Unspent_UnspentStatus `protobuf:"varint,11,opt,name=status,proto3,enum=types.Unspent_UnspentStatus" json:"status,omitempty"`
}

func (x *Unspent) Reset() {
	*x = Unspent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_types_types_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Unspent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Unspent) ProtoMessage() {}

func (x *Unspent) ProtoReflect() protoreflect.Message {
	mi := &file_types_types_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Unspent.ProtoReflect.Descriptor instead.
func (*Unspent) Descriptor() ([]byte, []int) {
	return file_types_types_proto_rawDescGZIP(), []int{2}
}

func (x *Unspent) GetTxid() string {
	if x != nil {
		return x.Txid
	}
	return ""
}

func (x *Unspent) GetIndex() int64 {
	if x != nil {
		return x.Index
	}
	return 0
}

func (x *Unspent) GetAsset() []byte {
	if x != nil {
		return x.Asset
	}
	return nil
}

func (x *Unspent) GetAssetCommitment() []byte {
	if x != nil {
		return x.AssetCommitment
	}
	return nil
}

func (x *Unspent) GetValue() []byte {
	if x != nil {
		return x.Value
	}
	return nil
}

func (x *Unspent) GetValueCommitment() []byte {
	if x != nil {
		return x.ValueCommitment
	}
	return nil
}

func (x *Unspent) GetScript() []byte {
	if x != nil {
		return x.Script
	}
	return nil
}

func (x *Unspent) GetNonce() []byte {
	if x != nil {
		return x.Nonce
	}
	return nil
}

func (x *Unspent) GetRangeProof() []byte {
	if x != nil {
		return x.RangeProof
	}
	return nil
}

func (x *Unspent) GetSurjectionProof() []byte {
	if x != nil {
		return x.SurjectionProof
	}
	return nil
}

func (x *Unspent) GetStatus() Unspent_UnspentStatus {
	if x != nil {
		return x.Status
	}
	return Unspent_CONFIRMED
}

type Template struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Format Template_Format `protobuf:"varint,1,opt,name=format,proto3,enum=types.Template_Format" json:"format,omitempty"`
	Value  string          `protobuf:"bytes,2,opt,name=value,proto3" json:"value,omitempty"`
}

func (x *Template) Reset() {
	*x = Template{}
	if protoimpl.UnsafeEnabled {
		mi := &file_types_types_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Template) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Template) ProtoMessage() {}

func (x *Template) ProtoReflect() protoreflect.Message {
	mi := &file_types_types_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Template.ProtoReflect.Descriptor instead.
func (*Template) Descriptor() ([]byte, []int) {
	return file_types_types_proto_rawDescGZIP(), []int{3}
}

func (x *Template) GetFormat() Template_Format {
	if x != nil {
		return x.Format
	}
	return Template_DESCRIPTOR
}

func (x *Template) GetValue() string {
	if x != nil {
		return x.Value
	}
	return ""
}

var File_types_types_proto protoreflect.FileDescriptor

var file_types_types_proto_rawDesc = []byte{
	0x0a, 0x11, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x05, 0x74, 0x79, 0x70, 0x65, 0x73, 0x22, 0x30, 0x0a, 0x0a, 0x41, 0x63,
	0x63, 0x6f, 0x75, 0x6e, 0x74, 0x4b, 0x65, 0x79, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x02, 0x69, 0x64, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x6a, 0x0a, 0x08,
	0x55, 0x6e, 0x73, 0x70, 0x65, 0x6e, 0x74, 0x73, 0x12, 0x32, 0x0a, 0x0b, 0x61, 0x63, 0x63, 0x6f,
	0x75, 0x6e, 0x74, 0x5f, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x11, 0x2e,
	0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x41, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x4b, 0x65, 0x79,
	0x52, 0x0a, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x4b, 0x65, 0x79, 0x12, 0x2a, 0x0a, 0x08,
	0x75, 0x6e, 0x73, 0x70, 0x65, 0x6e, 0x74, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0e,
	0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x55, 0x6e, 0x73, 0x70, 0x65, 0x6e, 0x74, 0x52, 0x08,
	0x75, 0x6e, 0x73, 0x70, 0x65, 0x6e, 0x74, 0x73, 0x22, 0xae, 0x03, 0x0a, 0x07, 0x55, 0x6e, 0x73,
	0x70, 0x65, 0x6e, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x78, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x74, 0x78, 0x69, 0x64, 0x12, 0x14, 0x0a, 0x05, 0x69, 0x6e, 0x64, 0x65,
	0x78, 0x18, 0x02, 0x20, 0x01, 0x28, 0x03, 0x52, 0x05, 0x69, 0x6e, 0x64, 0x65, 0x78, 0x12, 0x14,
	0x0a, 0x05, 0x61, 0x73, 0x73, 0x65, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x61,
	0x73, 0x73, 0x65, 0x74, 0x12, 0x29, 0x0a, 0x10, 0x61, 0x73, 0x73, 0x65, 0x74, 0x5f, 0x63, 0x6f,
	0x6d, 0x6d, 0x69, 0x74, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0f,
	0x61, 0x73, 0x73, 0x65, 0x74, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x6d, 0x65, 0x6e, 0x74, 0x12,
	0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05,
	0x76, 0x61, 0x6c, 0x75, 0x65, 0x12, 0x29, 0x0a, 0x10, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x5f, 0x63,
	0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x6d, 0x65, 0x6e, 0x74, 0x18, 0x06, 0x20, 0x01, 0x28, 0x0c, 0x52,
	0x0f, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x43, 0x6f, 0x6d, 0x6d, 0x69, 0x74, 0x6d, 0x65, 0x6e, 0x74,
	0x12, 0x16, 0x0a, 0x06, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0c,
	0x52, 0x06, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x6e, 0x6f, 0x6e, 0x63,
	0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x05, 0x6e, 0x6f, 0x6e, 0x63, 0x65, 0x12, 0x1e,
	0x0a, 0x0a, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x50, 0x72, 0x6f, 0x6f, 0x66, 0x18, 0x09, 0x20, 0x01,
	0x28, 0x0c, 0x52, 0x0a, 0x72, 0x61, 0x6e, 0x67, 0x65, 0x50, 0x72, 0x6f, 0x6f, 0x66, 0x12, 0x28,
	0x0a, 0x0f, 0x73, 0x75, 0x72, 0x6a, 0x65, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x72, 0x6f, 0x6f,
	0x66, 0x18, 0x0a, 0x20, 0x01, 0x28, 0x0c, 0x52, 0x0f, 0x73, 0x75, 0x72, 0x6a, 0x65, 0x63, 0x74,
	0x69, 0x6f, 0x6e, 0x50, 0x72, 0x6f, 0x6f, 0x66, 0x12, 0x34, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x18, 0x0b, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1c, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73,
	0x2e, 0x55, 0x6e, 0x73, 0x70, 0x65, 0x6e, 0x74, 0x2e, 0x55, 0x6e, 0x73, 0x70, 0x65, 0x6e, 0x74,
	0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x22, 0x49,
	0x0a, 0x0d, 0x55, 0x6e, 0x73, 0x70, 0x65, 0x6e, 0x74, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12,
	0x0d, 0x0a, 0x09, 0x43, 0x4f, 0x4e, 0x46, 0x49, 0x52, 0x4d, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0f,
	0x0a, 0x0b, 0x55, 0x4e, 0x43, 0x4f, 0x4e, 0x46, 0x49, 0x52, 0x4d, 0x45, 0x44, 0x10, 0x01, 0x12,
	0x0a, 0x0a, 0x06, 0x4c, 0x4f, 0x43, 0x4b, 0x45, 0x44, 0x10, 0x02, 0x12, 0x0c, 0x0a, 0x08, 0x55,
	0x4e, 0x4c, 0x4f, 0x43, 0x4b, 0x45, 0x44, 0x10, 0x03, 0x22, 0x8e, 0x01, 0x0a, 0x08, 0x54, 0x65,
	0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x12, 0x2e, 0x0a, 0x06, 0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x16, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x54,
	0x65, 0x6d, 0x70, 0x6c, 0x61, 0x74, 0x65, 0x2e, 0x46, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x52, 0x06,
	0x66, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18,
	0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x22, 0x3c, 0x0a, 0x06,
	0x46, 0x6f, 0x72, 0x6d, 0x61, 0x74, 0x12, 0x0e, 0x0a, 0x0a, 0x44, 0x45, 0x53, 0x43, 0x52, 0x49,
	0x50, 0x54, 0x4f, 0x52, 0x10, 0x00, 0x12, 0x0e, 0x0a, 0x0a, 0x4d, 0x49, 0x4e, 0x49, 0x53, 0x43,
	0x52, 0x49, 0x50, 0x54, 0x10, 0x01, 0x12, 0x09, 0x0a, 0x05, 0x49, 0x4f, 0x4e, 0x49, 0x4f, 0x10,
	0x02, 0x12, 0x07, 0x0a, 0x03, 0x52, 0x41, 0x57, 0x10, 0x03, 0x2a, 0x3e, 0x0a, 0x0b, 0x54, 0x78,
	0x45, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0f, 0x0a, 0x0b, 0x42, 0x52, 0x4f,
	0x41, 0x44, 0x43, 0x41, 0x53, 0x54, 0x45, 0x44, 0x10, 0x00, 0x12, 0x0f, 0x0a, 0x0b, 0x55, 0x4e,
	0x43, 0x4f, 0x4e, 0x46, 0x49, 0x52, 0x4d, 0x45, 0x44, 0x10, 0x02, 0x12, 0x0d, 0x0a, 0x09, 0x43,
	0x4f, 0x4e, 0x46, 0x49, 0x52, 0x4d, 0x45, 0x44, 0x10, 0x03, 0x42, 0x88, 0x01, 0x0a, 0x09, 0x63,
	0x6f, 0x6d, 0x2e, 0x74, 0x79, 0x70, 0x65, 0x73, 0x42, 0x0a, 0x54, 0x79, 0x70, 0x65, 0x73, 0x50,
	0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01, 0x5a, 0x3b, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x76, 0x75, 0x6c, 0x70, 0x65, 0x6d, 0x76, 0x65, 0x6e, 0x74, 0x75, 0x72, 0x65,
	0x73, 0x2f, 0x6f, 0x63, 0x65, 0x61, 0x6e, 0x2f, 0x61, 0x70, 0x69, 0x2d, 0x73, 0x70, 0x65, 0x63,
	0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x74, 0x79,
	0x70, 0x65, 0x73, 0xa2, 0x02, 0x03, 0x54, 0x58, 0x58, 0xaa, 0x02, 0x05, 0x54, 0x79, 0x70, 0x65,
	0x73, 0xca, 0x02, 0x05, 0x54, 0x79, 0x70, 0x65, 0x73, 0xe2, 0x02, 0x11, 0x54, 0x79, 0x70, 0x65,
	0x73, 0x5c, 0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x05,
	0x54, 0x79, 0x70, 0x65, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_types_types_proto_rawDescOnce sync.Once
	file_types_types_proto_rawDescData = file_types_types_proto_rawDesc
)

func file_types_types_proto_rawDescGZIP() []byte {
	file_types_types_proto_rawDescOnce.Do(func() {
		file_types_types_proto_rawDescData = protoimpl.X.CompressGZIP(file_types_types_proto_rawDescData)
	})
	return file_types_types_proto_rawDescData
}

var file_types_types_proto_enumTypes = make([]protoimpl.EnumInfo, 3)
var file_types_types_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_types_types_proto_goTypes = []interface{}{
	(TxEventType)(0),           // 0: types.TxEventType
	(Unspent_UnspentStatus)(0), // 1: types.Unspent.UnspentStatus
	(Template_Format)(0),       // 2: types.Template.Format
	(*AccountKey)(nil),         // 3: types.AccountKey
	(*Unspents)(nil),           // 4: types.Unspents
	(*Unspent)(nil),            // 5: types.Unspent
	(*Template)(nil),           // 6: types.Template
}
var file_types_types_proto_depIdxs = []int32{
	3, // 0: types.Unspents.account_key:type_name -> types.AccountKey
	5, // 1: types.Unspents.unspents:type_name -> types.Unspent
	1, // 2: types.Unspent.status:type_name -> types.Unspent.UnspentStatus
	2, // 3: types.Template.format:type_name -> types.Template.Format
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_types_types_proto_init() }
func file_types_types_proto_init() {
	if File_types_types_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_types_types_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AccountKey); i {
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
		file_types_types_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Unspents); i {
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
		file_types_types_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Unspent); i {
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
		file_types_types_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Template); i {
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
			RawDescriptor: file_types_types_proto_rawDesc,
			NumEnums:      3,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_types_types_proto_goTypes,
		DependencyIndexes: file_types_types_proto_depIdxs,
		EnumInfos:         file_types_types_proto_enumTypes,
		MessageInfos:      file_types_types_proto_msgTypes,
	}.Build()
	File_types_types_proto = out.File
	file_types_types_proto_rawDesc = nil
	file_types_types_proto_goTypes = nil
	file_types_types_proto_depIdxs = nil
}
