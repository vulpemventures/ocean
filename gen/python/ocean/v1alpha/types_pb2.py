# -*- coding: utf-8 -*-
# Generated by the protocol buffer compiler.  DO NOT EDIT!
# source: ocean/v1alpha/types.proto
"""Generated protocol buffer code."""
from google.protobuf.internal import enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import descriptor_pool as _descriptor_pool
from google.protobuf import message as _message
from google.protobuf import reflection as _reflection
from google.protobuf import symbol_database as _symbol_database
# @@protoc_insertion_point(imports)

_sym_db = _symbol_database.Default()




DESCRIPTOR = _descriptor_pool.Default().AddSerializedFile(b'\n\x19ocean/v1alpha/types.proto\x12\rocean.v1alpha\"0\n\nAccountKey\x12\x0e\n\x02id\x18\x01 \x01(\x04R\x02id\x12\x12\n\x04name\x18\x02 \x01(\tR\x04name\"\x86\x01\n\x0b\x41\x63\x63ountInfo\x12:\n\x0b\x61\x63\x63ount_key\x18\x01 \x01(\x0b\x32\x19.ocean.v1alpha.AccountKeyR\naccountKey\x12\'\n\x0f\x64\x65rivation_path\x18\x02 \x01(\tR\x0e\x64\x65rivationPath\x12\x12\n\x04xpub\x18\x03 \x01(\tR\x04xpub\"\x90\x01\n\x0b\x42\x61lanceInfo\x12#\n\rtotal_balance\x18\x01 \x01(\x04R\x0ctotalBalance\x12+\n\x11\x63onfirmed_balance\x18\x02 \x01(\x04R\x10\x63onfirmedBalance\x12/\n\x13unconfirmed_balance\x18\x03 \x01(\x04R\x12unconfirmedBalance\"1\n\x05Input\x12\x12\n\x04txid\x18\x01 \x01(\tR\x04txid\x12\x14\n\x05index\x18\x02 \x01(\x03R\x05index\"P\n\x06Output\x12\x14\n\x05\x61sset\x18\x01 \x01(\tR\x05\x61sset\x12\x16\n\x06\x61mount\x18\x02 \x01(\x03R\x06\x61mount\x12\x18\n\x07\x61\x64\x64ress\x18\x03 \x01(\tR\x07\x61\x64\x64ress\"n\n\x05Utxos\x12:\n\x0b\x61\x63\x63ount_key\x18\x01 \x01(\x0b\x32\x19.ocean.v1alpha.AccountKeyR\naccountKey\x12)\n\x05utxos\x18\x02 \x03(\x0b\x32\x13.ocean.v1alpha.UtxoR\x05utxos\"\xb1\x01\n\rUtxoWithEvent\x12:\n\x0b\x61\x63\x63ount_key\x18\x01 \x01(\x0b\x32\x19.ocean.v1alpha.AccountKeyR\naccountKey\x12\'\n\x04utxo\x18\x02 \x01(\x0b\x32\x13.ocean.v1alpha.UtxoR\x04utxo\x12;\n\nevent_type\x18\x03 \x01(\x0e\x32\x1c.ocean.v1alpha.UtxoEventTypeR\teventType\"\xb4\x01\n\x04Utxo\x12\x12\n\x04txid\x18\x01 \x01(\tR\x04txid\x12\x14\n\x05index\x18\x02 \x01(\x03R\x05index\x12\x14\n\x05\x61sset\x18\x03 \x01(\x0cR\x05\x61sset\x12\x14\n\x05value\x18\x04 \x01(\x0cR\x05value\x12\x16\n\x06script\x18\x05 \x01(\x0cR\x06script\x12!\n\x0cis_confirmed\x18\x06 \x01(\x08R\x0bisConfirmed\x12\x1b\n\tis_locked\x18\x07 \x01(\x08R\x08isLocked\"\xca\x01\n\x08Template\x12\x36\n\x06\x66ormat\x18\x01 \x01(\x0e\x32\x1e.ocean.v1alpha.Template.FormatR\x06\x66ormat\x12\x14\n\x05value\x18\x02 \x01(\tR\x05value\"p\n\x06\x46ormat\x12\x16\n\x12\x46ORMAT_UNSPECIFIED\x10\x00\x12\x15\n\x11\x46ORMAT_DESCRIPTOR\x10\x01\x12\x15\n\x11\x46ORMAT_MINISCRIPT\x10\x02\x12\x10\n\x0c\x46ORMAT_IONIO\x10\x03\x12\x0e\n\nFORMAT_RAW\x10\x04*\x87\x01\n\x0bTxEventType\x12\x1d\n\x19TX_EVENT_TYPE_UNSPECIFIED\x10\x00\x12\x1d\n\x19TX_EVENT_TYPE_BROADCASTED\x10\x01\x12\x1d\n\x19TX_EVENT_TYPE_UNCONFIRMED\x10\x02\x12\x1b\n\x17TX_EVENT_TYPE_CONFIRMED\x10\x03*\x85\x01\n\rUtxoEventType\x12\x1f\n\x1bUTXO_EVENT_TYPE_UNSPECIFIED\x10\x00\x12\x1a\n\x16UTXO_EVENT_TYPE_LOCKED\x10\x01\x12\x1c\n\x18UTXO_EVENT_TYPE_UNLOCKED\x10\x02\x12\x19\n\x15UTXO_EVENT_TYPE_SPENT\x10\x03*{\n\x10WebhookEventType\x12\"\n\x1eWEBHOOK_EVENT_TYPE_UNSPECIFIED\x10\x00\x12\"\n\x1eWEBHOOK_EVENT_TYPE_TRANSACTION\x10\x01\x12\x1f\n\x1bWEBHOOK_EVENT_TYPE_UNSPENTS\x10\x02\x42\xd7\x01\n\x11\x63om.ocean.v1alphaB\nTypesProtoP\x01Zagithub.com/vulpemventures/ocean/api-spec/protobuf/ocean/v1alpha/gen/go/ocean/v1alpha;oceanv1alpha\xa2\x02\x03OXX\xaa\x02\rOcean.V1alpha\xca\x02\rOcean\\V1alpha\xe2\x02\x19Ocean\\V1alpha\\GPBMetadata\xea\x02\x0eOcean::V1alphab\x06proto3')

_TXEVENTTYPE = DESCRIPTOR.enum_types_by_name['TxEventType']
TxEventType = enum_type_wrapper.EnumTypeWrapper(_TXEVENTTYPE)
_UTXOEVENTTYPE = DESCRIPTOR.enum_types_by_name['UtxoEventType']
UtxoEventType = enum_type_wrapper.EnumTypeWrapper(_UTXOEVENTTYPE)
_WEBHOOKEVENTTYPE = DESCRIPTOR.enum_types_by_name['WebhookEventType']
WebhookEventType = enum_type_wrapper.EnumTypeWrapper(_WEBHOOKEVENTTYPE)
TX_EVENT_TYPE_UNSPECIFIED = 0
TX_EVENT_TYPE_BROADCASTED = 1
TX_EVENT_TYPE_UNCONFIRMED = 2
TX_EVENT_TYPE_CONFIRMED = 3
UTXO_EVENT_TYPE_UNSPECIFIED = 0
UTXO_EVENT_TYPE_LOCKED = 1
UTXO_EVENT_TYPE_UNLOCKED = 2
UTXO_EVENT_TYPE_SPENT = 3
WEBHOOK_EVENT_TYPE_UNSPECIFIED = 0
WEBHOOK_EVENT_TYPE_TRANSACTION = 1
WEBHOOK_EVENT_TYPE_UNSPENTS = 2


_ACCOUNTKEY = DESCRIPTOR.message_types_by_name['AccountKey']
_ACCOUNTINFO = DESCRIPTOR.message_types_by_name['AccountInfo']
_BALANCEINFO = DESCRIPTOR.message_types_by_name['BalanceInfo']
_INPUT = DESCRIPTOR.message_types_by_name['Input']
_OUTPUT = DESCRIPTOR.message_types_by_name['Output']
_UTXOS = DESCRIPTOR.message_types_by_name['Utxos']
_UTXOWITHEVENT = DESCRIPTOR.message_types_by_name['UtxoWithEvent']
_UTXO = DESCRIPTOR.message_types_by_name['Utxo']
_TEMPLATE = DESCRIPTOR.message_types_by_name['Template']
_TEMPLATE_FORMAT = _TEMPLATE.enum_types_by_name['Format']
AccountKey = _reflection.GeneratedProtocolMessageType('AccountKey', (_message.Message,), {
  'DESCRIPTOR' : _ACCOUNTKEY,
  '__module__' : 'ocean.v1alpha.types_pb2'
  # @@protoc_insertion_point(class_scope:ocean.v1alpha.AccountKey)
  })
_sym_db.RegisterMessage(AccountKey)

AccountInfo = _reflection.GeneratedProtocolMessageType('AccountInfo', (_message.Message,), {
  'DESCRIPTOR' : _ACCOUNTINFO,
  '__module__' : 'ocean.v1alpha.types_pb2'
  # @@protoc_insertion_point(class_scope:ocean.v1alpha.AccountInfo)
  })
_sym_db.RegisterMessage(AccountInfo)

BalanceInfo = _reflection.GeneratedProtocolMessageType('BalanceInfo', (_message.Message,), {
  'DESCRIPTOR' : _BALANCEINFO,
  '__module__' : 'ocean.v1alpha.types_pb2'
  # @@protoc_insertion_point(class_scope:ocean.v1alpha.BalanceInfo)
  })
_sym_db.RegisterMessage(BalanceInfo)

Input = _reflection.GeneratedProtocolMessageType('Input', (_message.Message,), {
  'DESCRIPTOR' : _INPUT,
  '__module__' : 'ocean.v1alpha.types_pb2'
  # @@protoc_insertion_point(class_scope:ocean.v1alpha.Input)
  })
_sym_db.RegisterMessage(Input)

Output = _reflection.GeneratedProtocolMessageType('Output', (_message.Message,), {
  'DESCRIPTOR' : _OUTPUT,
  '__module__' : 'ocean.v1alpha.types_pb2'
  # @@protoc_insertion_point(class_scope:ocean.v1alpha.Output)
  })
_sym_db.RegisterMessage(Output)

Utxos = _reflection.GeneratedProtocolMessageType('Utxos', (_message.Message,), {
  'DESCRIPTOR' : _UTXOS,
  '__module__' : 'ocean.v1alpha.types_pb2'
  # @@protoc_insertion_point(class_scope:ocean.v1alpha.Utxos)
  })
_sym_db.RegisterMessage(Utxos)

UtxoWithEvent = _reflection.GeneratedProtocolMessageType('UtxoWithEvent', (_message.Message,), {
  'DESCRIPTOR' : _UTXOWITHEVENT,
  '__module__' : 'ocean.v1alpha.types_pb2'
  # @@protoc_insertion_point(class_scope:ocean.v1alpha.UtxoWithEvent)
  })
_sym_db.RegisterMessage(UtxoWithEvent)

Utxo = _reflection.GeneratedProtocolMessageType('Utxo', (_message.Message,), {
  'DESCRIPTOR' : _UTXO,
  '__module__' : 'ocean.v1alpha.types_pb2'
  # @@protoc_insertion_point(class_scope:ocean.v1alpha.Utxo)
  })
_sym_db.RegisterMessage(Utxo)

Template = _reflection.GeneratedProtocolMessageType('Template', (_message.Message,), {
  'DESCRIPTOR' : _TEMPLATE,
  '__module__' : 'ocean.v1alpha.types_pb2'
  # @@protoc_insertion_point(class_scope:ocean.v1alpha.Template)
  })
_sym_db.RegisterMessage(Template)

if _descriptor._USE_C_DESCRIPTORS == False:

  DESCRIPTOR._options = None
  DESCRIPTOR._serialized_options = b'\n\021com.ocean.v1alphaB\nTypesProtoP\001Zagithub.com/vulpemventures/ocean/api-spec/protobuf/ocean/v1alpha/gen/go/ocean/v1alpha;oceanv1alpha\242\002\003OXX\252\002\rOcean.V1alpha\312\002\rOcean\\V1alpha\342\002\031Ocean\\V1alpha\\GPBMetadata\352\002\016Ocean::V1alpha'
  _TXEVENTTYPE._serialized_start=1192
  _TXEVENTTYPE._serialized_end=1327
  _UTXOEVENTTYPE._serialized_start=1330
  _UTXOEVENTTYPE._serialized_end=1463
  _WEBHOOKEVENTTYPE._serialized_start=1465
  _WEBHOOKEVENTTYPE._serialized_end=1588
  _ACCOUNTKEY._serialized_start=44
  _ACCOUNTKEY._serialized_end=92
  _ACCOUNTINFO._serialized_start=95
  _ACCOUNTINFO._serialized_end=229
  _BALANCEINFO._serialized_start=232
  _BALANCEINFO._serialized_end=376
  _INPUT._serialized_start=378
  _INPUT._serialized_end=427
  _OUTPUT._serialized_start=429
  _OUTPUT._serialized_end=509
  _UTXOS._serialized_start=511
  _UTXOS._serialized_end=621
  _UTXOWITHEVENT._serialized_start=624
  _UTXOWITHEVENT._serialized_end=801
  _UTXO._serialized_start=804
  _UTXO._serialized_end=984
  _TEMPLATE._serialized_start=987
  _TEMPLATE._serialized_end=1189
  _TEMPLATE_FORMAT._serialized_start=1077
  _TEMPLATE_FORMAT._serialized_end=1189
# @@protoc_insertion_point(module_scope)
