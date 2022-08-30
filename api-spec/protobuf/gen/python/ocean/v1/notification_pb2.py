# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: ocean/v1/notification.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import builder as _builder
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()


from ocean.v1 import types_pb2 as ocean_dot_v1_dot_types__pb2


DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x1bocean/v1/notification.proto\x12\x08ocean.v1\x1a\x14ocean/v1/types.proto\"!\n\x1fTransactionNotificationsRequest\"\xce\x01\n TransactionNotificationsResponse\x12#\n\raccount_names\x18\x01 \x03(\tR\x0c\x61\x63\x63ountNames\x12\x12\n\x04txid\x18\x02 \x01(\tR\x04txid\x12\x34\n\nevent_type\x18\x03 \x01(\x0e\x32\x15.ocean.v1.TxEventTypeR\teventType\x12;\n\rblock_details\x18\x04 \x01(\x0b\x32\x16.ocean.v1.BlockDetailsR\x0c\x62lockDetails\"\x1b\n\x19UtxosNotificationsRequest\"z\n\x1aUtxosNotificationsResponse\x12$\n\x05utxos\x18\x01 \x03(\x0b\x32\x0e.ocean.v1.UtxoR\x05utxos\x12\x36\n\nevent_type\x18\x02 \x01(\x0e\x32\x17.ocean.v1.UtxoEventTypeR\teventType\"\x82\x01\n\x11\x41\x64\x64WebhookRequest\x12\x1a\n\x08\x65ndpoint\x18\x01 \x01(\tR\x08\x65ndpoint\x12\x39\n\nevent_type\x18\x02 \x01(\x0e\x32\x1a.ocean.v1.WebhookEventTypeR\teventType\x12\x16\n\x06secret\x18\x03 \x01(\tR\x06secret\"$\n\x12\x41\x64\x64WebhookResponse\x12\x0e\n\x02id\x18\x01 \x01(\tR\x02id\"&\n\x14RemoveWebhookRequest\x12\x0e\n\x02id\x18\x01 \x01(\tR\x02id\"\x17\n\x15RemoveWebhookResponse\"P\n\x13ListWebhooksRequest\x12\x39\n\nevent_type\x18\x01 \x01(\x0e\x32\x1a.ocean.v1.WebhookEventTypeR\teventType\"P\n\x14ListWebhooksResponse\x12\x38\n\x0cwebhook_info\x18\x01 \x03(\x0b\x32\x15.ocean.v1.WebhookInfoR\x0bwebhookInfo\"X\n\x0bWebhookInfo\x12\x0e\n\x02id\x18\x01 \x01(\tR\x02id\x12\x1a\n\x08\x65ndpoint\x18\x02 \x01(\tR\x08\x65ndpoint\x12\x1d\n\nis_secured\x18\x03 \x01(\x08R\tisSecured2\xdd\x03\n\x13NotificationService\x12s\n\x18TransactionNotifications\x12).ocean.v1.TransactionNotificationsRequest\x1a*.ocean.v1.TransactionNotificationsResponse0\x01\x12\x61\n\x12UtxosNotifications\x12#.ocean.v1.UtxosNotificationsRequest\x1a$.ocean.v1.UtxosNotificationsResponse0\x01\x12I\n\nAddWebhook\x12\x1b.ocean.v1.AddWebhookRequest\x1a\x1c.ocean.v1.AddWebhookResponse\"\x00\x12R\n\rRemoveWebhook\x12\x1e.ocean.v1.RemoveWebhookRequest\x1a\x1f.ocean.v1.RemoveWebhookResponse\"\x00\x12O\n\x0cListWebhooks\x12\x1d.ocean.v1.ListWebhooksRequest\x1a\x1e.ocean.v1.ListWebhooksResponse\"\x00\x42\xaa\x01\n\x0c\x63om.ocean.v1B\x11NotificationProtoP\x01ZFgithub.com/vulpemventures/ocean/api-spec/protobuf/gen/ocean/v1;oceanv1\xa2\x02\x03OXX\xaa\x02\x08Ocean.V1\xca\x02\x08Ocean\\V1\xe2\x02\x14Ocean\\V1\\GPBMetadata\xea\x02\tOcean::V1b\x06proto3')

_builder.BuildMessageAndEnumDescriptors(DESCRIPTOR, globals())
_builder.BuildTopDescriptorsAndMessages(DESCRIPTOR, 'ocean.v1.notification_pb2', globals())
if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n\014com.ocean.v1B\021NotificationProtoP\001ZFgithub.com/vulpemventures/ocean/api-spec/protobuf/gen/ocean/v1;oceanv1\242\002\003OXX\252\002\010Ocean.V1\312\002\010Ocean\\V1\342\002\024Ocean\\V1\\GPBMetadata\352\002\tOcean::V1'
  _TRANSACTIONNOTIFICATIONSREQUEST._serialized_start=63
  _TRANSACTIONNOTIFICATIONSREQUEST._serialized_end=96
  _TRANSACTIONNOTIFICATIONSRESPONSE._serialized_start=99
  _TRANSACTIONNOTIFICATIONSRESPONSE._serialized_end=305
  _UTXOSNOTIFICATIONSREQUEST._serialized_start=307
  _UTXOSNOTIFICATIONSREQUEST._serialized_end=334
  _UTXOSNOTIFICATIONSRESPONSE._serialized_start=336
  _UTXOSNOTIFICATIONSRESPONSE._serialized_end=458
  _ADDWEBHOOKREQUEST._serialized_start=461
  _ADDWEBHOOKREQUEST._serialized_end=591
  _ADDWEBHOOKRESPONSE._serialized_start=593
  _ADDWEBHOOKRESPONSE._serialized_end=629
  _REMOVEWEBHOOKREQUEST._serialized_start=631
  _REMOVEWEBHOOKREQUEST._serialized_end=669
  _REMOVEWEBHOOKRESPONSE._serialized_start=671
  _REMOVEWEBHOOKRESPONSE._serialized_end=694
  _LISTWEBHOOKSREQUEST._serialized_start=696
  _LISTWEBHOOKSREQUEST._serialized_end=776
  _LISTWEBHOOKSRESPONSE._serialized_start=778
  _LISTWEBHOOKSRESPONSE._serialized_end=858
  _WEBHOOKINFO._serialized_start=860
  _WEBHOOKINFO._serialized_end=948
  _NOTIFICATIONSERVICE._serialized_start=951
  _NOTIFICATIONSERVICE._serialized_end=1428
# @@protoc_insertion_point(module_scope)