// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        (unknown)
// source: ocean/v1/notification.proto

package oceanv1

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

type TransactionNotificationsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *TransactionNotificationsRequest) Reset() {
	*x = TransactionNotificationsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ocean_v1_notification_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TransactionNotificationsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TransactionNotificationsRequest) ProtoMessage() {}

func (x *TransactionNotificationsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ocean_v1_notification_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TransactionNotificationsRequest.ProtoReflect.Descriptor instead.
func (*TransactionNotificationsRequest) Descriptor() ([]byte, []int) {
	return file_ocean_v1_notification_proto_rawDescGZIP(), []int{0}
}

type TransactionNotificationsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Tx event type.
	EventType TxEventType `protobuf:"varint,1,opt,name=event_type,json=eventType,proto3,enum=ocean.v1.TxEventType" json:"event_type,omitempty"`
	// Account names.
	AccountNames []string `protobuf:"bytes,2,rep,name=account_names,json=accountNames,proto3" json:"account_names,omitempty"`
	// Tx in hex format.
	Txhex string `protobuf:"bytes,3,opt,name=txhex,proto3" json:"txhex,omitempty"`
	// Txid of transaction.
	Txid string `protobuf:"bytes,4,opt,name=txid,proto3" json:"txid,omitempty"`
	// Details of the block including the tx.
	BlockDetails *BlockDetails `protobuf:"bytes,5,opt,name=block_details,json=blockDetails,proto3" json:"block_details,omitempty"`
}

func (x *TransactionNotificationsResponse) Reset() {
	*x = TransactionNotificationsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ocean_v1_notification_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *TransactionNotificationsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*TransactionNotificationsResponse) ProtoMessage() {}

func (x *TransactionNotificationsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ocean_v1_notification_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use TransactionNotificationsResponse.ProtoReflect.Descriptor instead.
func (*TransactionNotificationsResponse) Descriptor() ([]byte, []int) {
	return file_ocean_v1_notification_proto_rawDescGZIP(), []int{1}
}

func (x *TransactionNotificationsResponse) GetEventType() TxEventType {
	if x != nil {
		return x.EventType
	}
	return TxEventType_TX_EVENT_TYPE_UNSPECIFIED
}

func (x *TransactionNotificationsResponse) GetAccountNames() []string {
	if x != nil {
		return x.AccountNames
	}
	return nil
}

func (x *TransactionNotificationsResponse) GetTxhex() string {
	if x != nil {
		return x.Txhex
	}
	return ""
}

func (x *TransactionNotificationsResponse) GetTxid() string {
	if x != nil {
		return x.Txid
	}
	return ""
}

func (x *TransactionNotificationsResponse) GetBlockDetails() *BlockDetails {
	if x != nil {
		return x.BlockDetails
	}
	return nil
}

type UtxosNotificationsRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *UtxosNotificationsRequest) Reset() {
	*x = UtxosNotificationsRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ocean_v1_notification_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UtxosNotificationsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UtxosNotificationsRequest) ProtoMessage() {}

func (x *UtxosNotificationsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ocean_v1_notification_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UtxosNotificationsRequest.ProtoReflect.Descriptor instead.
func (*UtxosNotificationsRequest) Descriptor() ([]byte, []int) {
	return file_ocean_v1_notification_proto_rawDescGZIP(), []int{2}
}

type UtxosNotificationsResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The event's type occured for the utxos.
	EventType UtxoEventType `protobuf:"varint,1,opt,name=event_type,json=eventType,proto3,enum=ocean.v1.UtxoEventType" json:"event_type,omitempty"`
	// List of utxos for which occured the event.
	Utxos []*Utxo `protobuf:"bytes,2,rep,name=utxos,proto3" json:"utxos,omitempty"`
}

func (x *UtxosNotificationsResponse) Reset() {
	*x = UtxosNotificationsResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ocean_v1_notification_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *UtxosNotificationsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*UtxosNotificationsResponse) ProtoMessage() {}

func (x *UtxosNotificationsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ocean_v1_notification_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use UtxosNotificationsResponse.ProtoReflect.Descriptor instead.
func (*UtxosNotificationsResponse) Descriptor() ([]byte, []int) {
	return file_ocean_v1_notification_proto_rawDescGZIP(), []int{3}
}

func (x *UtxosNotificationsResponse) GetEventType() UtxoEventType {
	if x != nil {
		return x.EventType
	}
	return UtxoEventType_UTXO_EVENT_TYPE_UNSPECIFIED
}

func (x *UtxosNotificationsResponse) GetUtxos() []*Utxo {
	if x != nil {
		return x.Utxos
	}
	return nil
}

type AddWebhookRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The endpoint of the external service to reach.
	Endpoint string `protobuf:"bytes,1,opt,name=endpoint,proto3" json:"endpoint,omitempty"`
	// The event type for which the webhook should be registered.
	EventType WebhookEventType `protobuf:"varint,2,opt,name=event_type,json=eventType,proto3,enum=ocean.v1.WebhookEventType" json:"event_type,omitempty"`
	// The secret to use for signign a JWT token for an authenticated request
	// to the external service.
	Secret string `protobuf:"bytes,3,opt,name=secret,proto3" json:"secret,omitempty"`
}

func (x *AddWebhookRequest) Reset() {
	*x = AddWebhookRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ocean_v1_notification_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddWebhookRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddWebhookRequest) ProtoMessage() {}

func (x *AddWebhookRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ocean_v1_notification_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddWebhookRequest.ProtoReflect.Descriptor instead.
func (*AddWebhookRequest) Descriptor() ([]byte, []int) {
	return file_ocean_v1_notification_proto_rawDescGZIP(), []int{4}
}

func (x *AddWebhookRequest) GetEndpoint() string {
	if x != nil {
		return x.Endpoint
	}
	return ""
}

func (x *AddWebhookRequest) GetEventType() WebhookEventType {
	if x != nil {
		return x.EventType
	}
	return WebhookEventType_WEBHOOK_EVENT_TYPE_UNSPECIFIED
}

func (x *AddWebhookRequest) GetSecret() string {
	if x != nil {
		return x.Secret
	}
	return ""
}

type AddWebhookResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The id of the new webhook.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *AddWebhookResponse) Reset() {
	*x = AddWebhookResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ocean_v1_notification_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AddWebhookResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AddWebhookResponse) ProtoMessage() {}

func (x *AddWebhookResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ocean_v1_notification_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AddWebhookResponse.ProtoReflect.Descriptor instead.
func (*AddWebhookResponse) Descriptor() ([]byte, []int) {
	return file_ocean_v1_notification_proto_rawDescGZIP(), []int{5}
}

func (x *AddWebhookResponse) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type RemoveWebhookRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The id of the webhook to remove.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *RemoveWebhookRequest) Reset() {
	*x = RemoveWebhookRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ocean_v1_notification_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RemoveWebhookRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoveWebhookRequest) ProtoMessage() {}

func (x *RemoveWebhookRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ocean_v1_notification_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoveWebhookRequest.ProtoReflect.Descriptor instead.
func (*RemoveWebhookRequest) Descriptor() ([]byte, []int) {
	return file_ocean_v1_notification_proto_rawDescGZIP(), []int{6}
}

func (x *RemoveWebhookRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

type RemoveWebhookResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *RemoveWebhookResponse) Reset() {
	*x = RemoveWebhookResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ocean_v1_notification_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RemoveWebhookResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RemoveWebhookResponse) ProtoMessage() {}

func (x *RemoveWebhookResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ocean_v1_notification_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RemoveWebhookResponse.ProtoReflect.Descriptor instead.
func (*RemoveWebhookResponse) Descriptor() ([]byte, []int) {
	return file_ocean_v1_notification_proto_rawDescGZIP(), []int{7}
}

type ListWebhooksRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The event type for which filtering the list of webhooks.
	EventType WebhookEventType `protobuf:"varint,1,opt,name=event_type,json=eventType,proto3,enum=ocean.v1.WebhookEventType" json:"event_type,omitempty"`
}

func (x *ListWebhooksRequest) Reset() {
	*x = ListWebhooksRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ocean_v1_notification_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListWebhooksRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListWebhooksRequest) ProtoMessage() {}

func (x *ListWebhooksRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ocean_v1_notification_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListWebhooksRequest.ProtoReflect.Descriptor instead.
func (*ListWebhooksRequest) Descriptor() ([]byte, []int) {
	return file_ocean_v1_notification_proto_rawDescGZIP(), []int{8}
}

func (x *ListWebhooksRequest) GetEventType() WebhookEventType {
	if x != nil {
		return x.EventType
	}
	return WebhookEventType_WEBHOOK_EVENT_TYPE_UNSPECIFIED
}

type ListWebhooksResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The list of info about the webhooks regitered for an action.
	WebhookInfo []*WebhookInfo `protobuf:"bytes,1,rep,name=webhook_info,json=webhookInfo,proto3" json:"webhook_info,omitempty"`
}

func (x *ListWebhooksResponse) Reset() {
	*x = ListWebhooksResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ocean_v1_notification_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ListWebhooksResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListWebhooksResponse) ProtoMessage() {}

func (x *ListWebhooksResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ocean_v1_notification_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListWebhooksResponse.ProtoReflect.Descriptor instead.
func (*ListWebhooksResponse) Descriptor() ([]byte, []int) {
	return file_ocean_v1_notification_proto_rawDescGZIP(), []int{9}
}

func (x *ListWebhooksResponse) GetWebhookInfo() []*WebhookInfo {
	if x != nil {
		return x.WebhookInfo
	}
	return nil
}

type WebhookInfo struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// The id of the webhook.
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// The endpoint of the external service to reach.
	Endpoint string `protobuf:"bytes,2,opt,name=endpoint,proto3" json:"endpoint,omitempty"`
	// Whether the outgoing requests are authenticated.
	IsSecured bool `protobuf:"varint,3,opt,name=is_secured,json=isSecured,proto3" json:"is_secured,omitempty"`
}

func (x *WebhookInfo) Reset() {
	*x = WebhookInfo{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ocean_v1_notification_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *WebhookInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*WebhookInfo) ProtoMessage() {}

func (x *WebhookInfo) ProtoReflect() protoreflect.Message {
	mi := &file_ocean_v1_notification_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use WebhookInfo.ProtoReflect.Descriptor instead.
func (*WebhookInfo) Descriptor() ([]byte, []int) {
	return file_ocean_v1_notification_proto_rawDescGZIP(), []int{10}
}

func (x *WebhookInfo) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *WebhookInfo) GetEndpoint() string {
	if x != nil {
		return x.Endpoint
	}
	return ""
}

func (x *WebhookInfo) GetIsSecured() bool {
	if x != nil {
		return x.IsSecured
	}
	return false
}

var File_ocean_v1_notification_proto protoreflect.FileDescriptor

var file_ocean_v1_notification_proto_rawDesc = []byte{
	0x0a, 0x1b, 0x6f, 0x63, 0x65, 0x61, 0x6e, 0x2f, 0x76, 0x31, 0x2f, 0x6e, 0x6f, 0x74, 0x69, 0x66,
	0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x08, 0x6f,
	0x63, 0x65, 0x61, 0x6e, 0x2e, 0x76, 0x31, 0x1a, 0x14, 0x6f, 0x63, 0x65, 0x61, 0x6e, 0x2f, 0x76,
	0x31, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x21, 0x0a,
	0x1f, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4e, 0x6f, 0x74, 0x69,
	0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x22, 0xe4, 0x01, 0x0a, 0x20, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x34, 0x0a, 0x0a, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x74,
	0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x15, 0x2e, 0x6f, 0x63, 0x65, 0x61,
	0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x78, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65,
	0x52, 0x09, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x23, 0x0a, 0x0d, 0x61,
	0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x18, 0x02, 0x20, 0x03,
	0x28, 0x09, 0x52, 0x0c, 0x61, 0x63, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x73,
	0x12, 0x14, 0x0a, 0x05, 0x74, 0x78, 0x68, 0x65, 0x78, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x05, 0x74, 0x78, 0x68, 0x65, 0x78, 0x12, 0x12, 0x0a, 0x04, 0x74, 0x78, 0x69, 0x64, 0x18, 0x04,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x74, 0x78, 0x69, 0x64, 0x12, 0x3b, 0x0a, 0x0d, 0x62, 0x6c,
	0x6f, 0x63, 0x6b, 0x5f, 0x64, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x16, 0x2e, 0x6f, 0x63, 0x65, 0x61, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x42, 0x6c, 0x6f,
	0x63, 0x6b, 0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x52, 0x0c, 0x62, 0x6c, 0x6f, 0x63, 0x6b,
	0x44, 0x65, 0x74, 0x61, 0x69, 0x6c, 0x73, 0x22, 0x1b, 0x0a, 0x19, 0x55, 0x74, 0x78, 0x6f, 0x73,
	0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x22, 0x7a, 0x0a, 0x1a, 0x55, 0x74, 0x78, 0x6f, 0x73, 0x4e, 0x6f, 0x74,
	0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x12, 0x36, 0x0a, 0x0a, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x17, 0x2e, 0x6f, 0x63, 0x65, 0x61, 0x6e, 0x2e, 0x76,
	0x31, 0x2e, 0x55, 0x74, 0x78, 0x6f, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52,
	0x09, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x24, 0x0a, 0x05, 0x75, 0x74,
	0x78, 0x6f, 0x73, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x0e, 0x2e, 0x6f, 0x63, 0x65, 0x61,
	0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x74, 0x78, 0x6f, 0x52, 0x05, 0x75, 0x74, 0x78, 0x6f, 0x73,
	0x22, 0x82, 0x01, 0x0a, 0x11, 0x41, 0x64, 0x64, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x1a, 0x0a, 0x08, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69,
	0x6e, 0x74, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69,
	0x6e, 0x74, 0x12, 0x39, 0x0a, 0x0a, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70, 0x65,
	0x18, 0x02, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1a, 0x2e, 0x6f, 0x63, 0x65, 0x61, 0x6e, 0x2e, 0x76,
	0x31, 0x2e, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79,
	0x70, 0x65, 0x52, 0x09, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x16, 0x0a,
	0x06, 0x73, 0x65, 0x63, 0x72, 0x65, 0x74, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x73,
	0x65, 0x63, 0x72, 0x65, 0x74, 0x22, 0x24, 0x0a, 0x12, 0x41, 0x64, 0x64, 0x57, 0x65, 0x62, 0x68,
	0x6f, 0x6f, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69,
	0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x22, 0x26, 0x0a, 0x14, 0x52,
	0x65, 0x6d, 0x6f, 0x76, 0x65, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x02, 0x69, 0x64, 0x22, 0x17, 0x0a, 0x15, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x57, 0x65, 0x62,
	0x68, 0x6f, 0x6f, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x50, 0x0a, 0x13,
	0x4c, 0x69, 0x73, 0x74, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x73, 0x52, 0x65, 0x71, 0x75,
	0x65, 0x73, 0x74, 0x12, 0x39, 0x0a, 0x0a, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x74, 0x79, 0x70,
	0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x1a, 0x2e, 0x6f, 0x63, 0x65, 0x61, 0x6e, 0x2e,
	0x76, 0x31, 0x2e, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x54,
	0x79, 0x70, 0x65, 0x52, 0x09, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x22, 0x50,
	0x0a, 0x14, 0x4c, 0x69, 0x73, 0x74, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x73, 0x52, 0x65,
	0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x38, 0x0a, 0x0c, 0x77, 0x65, 0x62, 0x68, 0x6f, 0x6f,
	0x6b, 0x5f, 0x69, 0x6e, 0x66, 0x6f, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x15, 0x2e, 0x6f,
	0x63, 0x65, 0x61, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x49,
	0x6e, 0x66, 0x6f, 0x52, 0x0b, 0x77, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x49, 0x6e, 0x66, 0x6f,
	0x22, 0x58, 0x0a, 0x0b, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x49, 0x6e, 0x66, 0x6f, 0x12,
	0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12,
	0x1a, 0x0a, 0x08, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x65, 0x6e, 0x64, 0x70, 0x6f, 0x69, 0x6e, 0x74, 0x12, 0x1d, 0x0a, 0x0a, 0x69,
	0x73, 0x5f, 0x73, 0x65, 0x63, 0x75, 0x72, 0x65, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52,
	0x09, 0x69, 0x73, 0x53, 0x65, 0x63, 0x75, 0x72, 0x65, 0x64, 0x32, 0xdd, 0x03, 0x0a, 0x13, 0x4e,
	0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x12, 0x73, 0x0a, 0x18, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f,
	0x6e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x29,
	0x2e, 0x6f, 0x63, 0x65, 0x61, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61,
	0x63, 0x74, 0x69, 0x6f, 0x6e, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x2a, 0x2e, 0x6f, 0x63, 0x65, 0x61,
	0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x54, 0x72, 0x61, 0x6e, 0x73, 0x61, 0x63, 0x74, 0x69, 0x6f, 0x6e,
	0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x30, 0x01, 0x12, 0x61, 0x0a, 0x12, 0x55, 0x74, 0x78, 0x6f, 0x73,
	0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x12, 0x23, 0x2e,
	0x6f, 0x63, 0x65, 0x61, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x74, 0x78, 0x6f, 0x73, 0x4e, 0x6f,
	0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x24, 0x2e, 0x6f, 0x63, 0x65, 0x61, 0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x55, 0x74,
	0x78, 0x6f, 0x73, 0x4e, 0x6f, 0x74, 0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x30, 0x01, 0x12, 0x49, 0x0a, 0x0a, 0x41, 0x64,
	0x64, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x12, 0x1b, 0x2e, 0x6f, 0x63, 0x65, 0x61, 0x6e,
	0x2e, 0x76, 0x31, 0x2e, 0x41, 0x64, 0x64, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x52, 0x65,
	0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1c, 0x2e, 0x6f, 0x63, 0x65, 0x61, 0x6e, 0x2e, 0x76, 0x31,
	0x2e, 0x41, 0x64, 0x64, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x52, 0x0a, 0x0d, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x57,
	0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x12, 0x1e, 0x2e, 0x6f, 0x63, 0x65, 0x61, 0x6e, 0x2e, 0x76,
	0x31, 0x2e, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1f, 0x2e, 0x6f, 0x63, 0x65, 0x61, 0x6e, 0x2e, 0x76,
	0x31, 0x2e, 0x52, 0x65, 0x6d, 0x6f, 0x76, 0x65, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x52,
	0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x12, 0x4f, 0x0a, 0x0c, 0x4c, 0x69, 0x73,
	0x74, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x73, 0x12, 0x1d, 0x2e, 0x6f, 0x63, 0x65, 0x61,
	0x6e, 0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b,
	0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x1e, 0x2e, 0x6f, 0x63, 0x65, 0x61, 0x6e,
	0x2e, 0x76, 0x31, 0x2e, 0x4c, 0x69, 0x73, 0x74, 0x57, 0x65, 0x62, 0x68, 0x6f, 0x6f, 0x6b, 0x73,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0xaa, 0x01, 0x0a, 0x0c, 0x63,
	0x6f, 0x6d, 0x2e, 0x6f, 0x63, 0x65, 0x61, 0x6e, 0x2e, 0x76, 0x31, 0x42, 0x11, 0x4e, 0x6f, 0x74,
	0x69, 0x66, 0x69, 0x63, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x50, 0x72, 0x6f, 0x74, 0x6f, 0x50, 0x01,
	0x5a, 0x46, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x76, 0x75, 0x6c,
	0x70, 0x65, 0x6d, 0x76, 0x65, 0x6e, 0x74, 0x75, 0x72, 0x65, 0x73, 0x2f, 0x6f, 0x63, 0x65, 0x61,
	0x6e, 0x2f, 0x61, 0x70, 0x69, 0x2d, 0x73, 0x70, 0x65, 0x63, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2f, 0x67, 0x65, 0x6e, 0x2f, 0x6f, 0x63, 0x65, 0x61, 0x6e, 0x2f, 0x76, 0x31,
	0x3b, 0x6f, 0x63, 0x65, 0x61, 0x6e, 0x76, 0x31, 0xa2, 0x02, 0x03, 0x4f, 0x58, 0x58, 0xaa, 0x02,
	0x08, 0x4f, 0x63, 0x65, 0x61, 0x6e, 0x2e, 0x56, 0x31, 0xca, 0x02, 0x08, 0x4f, 0x63, 0x65, 0x61,
	0x6e, 0x5c, 0x56, 0x31, 0xe2, 0x02, 0x14, 0x4f, 0x63, 0x65, 0x61, 0x6e, 0x5c, 0x56, 0x31, 0x5c,
	0x47, 0x50, 0x42, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0xea, 0x02, 0x09, 0x4f, 0x63,
	0x65, 0x61, 0x6e, 0x3a, 0x3a, 0x56, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_ocean_v1_notification_proto_rawDescOnce sync.Once
	file_ocean_v1_notification_proto_rawDescData = file_ocean_v1_notification_proto_rawDesc
)

func file_ocean_v1_notification_proto_rawDescGZIP() []byte {
	file_ocean_v1_notification_proto_rawDescOnce.Do(func() {
		file_ocean_v1_notification_proto_rawDescData = protoimpl.X.CompressGZIP(file_ocean_v1_notification_proto_rawDescData)
	})
	return file_ocean_v1_notification_proto_rawDescData
}

var file_ocean_v1_notification_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_ocean_v1_notification_proto_goTypes = []interface{}{
	(*TransactionNotificationsRequest)(nil),  // 0: ocean.v1.TransactionNotificationsRequest
	(*TransactionNotificationsResponse)(nil), // 1: ocean.v1.TransactionNotificationsResponse
	(*UtxosNotificationsRequest)(nil),        // 2: ocean.v1.UtxosNotificationsRequest
	(*UtxosNotificationsResponse)(nil),       // 3: ocean.v1.UtxosNotificationsResponse
	(*AddWebhookRequest)(nil),                // 4: ocean.v1.AddWebhookRequest
	(*AddWebhookResponse)(nil),               // 5: ocean.v1.AddWebhookResponse
	(*RemoveWebhookRequest)(nil),             // 6: ocean.v1.RemoveWebhookRequest
	(*RemoveWebhookResponse)(nil),            // 7: ocean.v1.RemoveWebhookResponse
	(*ListWebhooksRequest)(nil),              // 8: ocean.v1.ListWebhooksRequest
	(*ListWebhooksResponse)(nil),             // 9: ocean.v1.ListWebhooksResponse
	(*WebhookInfo)(nil),                      // 10: ocean.v1.WebhookInfo
	(TxEventType)(0),                         // 11: ocean.v1.TxEventType
	(*BlockDetails)(nil),                     // 12: ocean.v1.BlockDetails
	(UtxoEventType)(0),                       // 13: ocean.v1.UtxoEventType
	(*Utxo)(nil),                             // 14: ocean.v1.Utxo
	(WebhookEventType)(0),                    // 15: ocean.v1.WebhookEventType
}
var file_ocean_v1_notification_proto_depIdxs = []int32{
	11, // 0: ocean.v1.TransactionNotificationsResponse.event_type:type_name -> ocean.v1.TxEventType
	12, // 1: ocean.v1.TransactionNotificationsResponse.block_details:type_name -> ocean.v1.BlockDetails
	13, // 2: ocean.v1.UtxosNotificationsResponse.event_type:type_name -> ocean.v1.UtxoEventType
	14, // 3: ocean.v1.UtxosNotificationsResponse.utxos:type_name -> ocean.v1.Utxo
	15, // 4: ocean.v1.AddWebhookRequest.event_type:type_name -> ocean.v1.WebhookEventType
	15, // 5: ocean.v1.ListWebhooksRequest.event_type:type_name -> ocean.v1.WebhookEventType
	10, // 6: ocean.v1.ListWebhooksResponse.webhook_info:type_name -> ocean.v1.WebhookInfo
	0,  // 7: ocean.v1.NotificationService.TransactionNotifications:input_type -> ocean.v1.TransactionNotificationsRequest
	2,  // 8: ocean.v1.NotificationService.UtxosNotifications:input_type -> ocean.v1.UtxosNotificationsRequest
	4,  // 9: ocean.v1.NotificationService.AddWebhook:input_type -> ocean.v1.AddWebhookRequest
	6,  // 10: ocean.v1.NotificationService.RemoveWebhook:input_type -> ocean.v1.RemoveWebhookRequest
	8,  // 11: ocean.v1.NotificationService.ListWebhooks:input_type -> ocean.v1.ListWebhooksRequest
	1,  // 12: ocean.v1.NotificationService.TransactionNotifications:output_type -> ocean.v1.TransactionNotificationsResponse
	3,  // 13: ocean.v1.NotificationService.UtxosNotifications:output_type -> ocean.v1.UtxosNotificationsResponse
	5,  // 14: ocean.v1.NotificationService.AddWebhook:output_type -> ocean.v1.AddWebhookResponse
	7,  // 15: ocean.v1.NotificationService.RemoveWebhook:output_type -> ocean.v1.RemoveWebhookResponse
	9,  // 16: ocean.v1.NotificationService.ListWebhooks:output_type -> ocean.v1.ListWebhooksResponse
	12, // [12:17] is the sub-list for method output_type
	7,  // [7:12] is the sub-list for method input_type
	7,  // [7:7] is the sub-list for extension type_name
	7,  // [7:7] is the sub-list for extension extendee
	0,  // [0:7] is the sub-list for field type_name
}

func init() { file_ocean_v1_notification_proto_init() }
func file_ocean_v1_notification_proto_init() {
	if File_ocean_v1_notification_proto != nil {
		return
	}
	file_ocean_v1_types_proto_init()
	if !protoimpl.UnsafeEnabled {
		file_ocean_v1_notification_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TransactionNotificationsRequest); i {
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
		file_ocean_v1_notification_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*TransactionNotificationsResponse); i {
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
		file_ocean_v1_notification_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UtxosNotificationsRequest); i {
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
		file_ocean_v1_notification_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*UtxosNotificationsResponse); i {
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
		file_ocean_v1_notification_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddWebhookRequest); i {
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
		file_ocean_v1_notification_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AddWebhookResponse); i {
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
		file_ocean_v1_notification_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RemoveWebhookRequest); i {
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
		file_ocean_v1_notification_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RemoveWebhookResponse); i {
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
		file_ocean_v1_notification_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListWebhooksRequest); i {
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
		file_ocean_v1_notification_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ListWebhooksResponse); i {
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
		file_ocean_v1_notification_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*WebhookInfo); i {
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
			RawDescriptor: file_ocean_v1_notification_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   11,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_ocean_v1_notification_proto_goTypes,
		DependencyIndexes: file_ocean_v1_notification_proto_depIdxs,
		MessageInfos:      file_ocean_v1_notification_proto_msgTypes,
	}.Build()
	File_ocean_v1_notification_proto = out.File
	file_ocean_v1_notification_proto_rawDesc = nil
	file_ocean_v1_notification_proto_goTypes = nil
	file_ocean_v1_notification_proto_depIdxs = nil
}
